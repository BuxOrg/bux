package bux

import (
	"context"

	"github.com/BuxOrg/bux/notifications"
	"github.com/mrz1836/go-datastore"
)

// GetModelTableName will get the db table name of the current model
func (m *Transaction) GetModelTableName() string {
	return tableTransactions
}

// Save will save the model into the Datastore
func (m *Transaction) Save(ctx context.Context) (err error) {
	// Prepare the metadata
	if len(m.Metadata) > 0 {
		// set the metadata to be xpub specific, but only if we have a valid xpub ID
		if m.XPubID != "" {
			// was metadata set via opts ?
			if m.XpubMetadata == nil {
				m.XpubMetadata = make(XpubMetadata)
			}
			if _, ok := m.XpubMetadata[m.XPubID]; !ok {
				m.XpubMetadata[m.XPubID] = make(Metadata)
			}
			for key, value := range m.Metadata {
				m.XpubMetadata[m.XPubID][key] = value
			}
		} else {
			m.DebugLog("xPub id is missing from transaction, cannot store metadata")
		}
	}

	return Save(ctx, m)
}

// BeforeCreating will fire before the model is being inserted into the Datastore
func (m *Transaction) BeforeCreating(ctx context.Context) error {
	if m.beforeCreateCalled {
		m.DebugLog("skipping: " + m.Name() + " BeforeCreating hook, because already called")
		return nil
	}

	m.DebugLog("starting: " + m.Name() + " BeforeCreating hook...")

	// Test for required field(s)
	if len(m.Hex) == 0 {
		return ErrMissingFieldHex
	}

	// Set the xPubID
	m.setXPubID()

	// Set the ID - will also parse and verify the tx
	err := m.setID()
	if err != nil {
		return err
	}

	// 	m.xPubID is the xpub of the user registering the transaction
	if len(m.XPubID) > 0 && len(m.DraftID) > 0 {
		// Only get the draft if we haven't already
		if m.draftTransaction == nil {
			if m.draftTransaction, err = getDraftTransactionID(
				ctx, m.XPubID, m.DraftID, m.GetOptions(false)...,
			); err != nil {
				return err
			} else if m.draftTransaction == nil {
				return ErrDraftNotFound
			}
		}
	}

	// Validations and broadcast config check
	if m.draftTransaction != nil {

		// No config set? Use the default from the client
		if m.draftTransaction.Configuration.Sync == nil {
			m.draftTransaction.Configuration.Sync = m.Client().DefaultSyncConfig()
		}

		// Create the sync transaction model
		sync := newSyncTransaction(
			m.GetID(),
			m.draftTransaction.Configuration.Sync,
			m.GetOptions(true)...,
		)

		// Found any p2p outputs?
		p2pStatus := SyncStatusSkipped
		if m.draftTransaction.Configuration.Outputs != nil {
			for _, output := range m.draftTransaction.Configuration.Outputs {
				if output.PaymailP4 != nil && output.PaymailP4.ResolutionType == ResolutionTypeP2P {
					p2pStatus = SyncStatusPending
				}
			}
		}
		sync.P2PStatus = p2pStatus

		// Use the same metadata
		sync.Metadata = m.Metadata

		// set this transaction on the sync transaction object. This is needed for the first broadcast
		sync.transaction = m

		// If all the options are skipped, do not make a new model (ignore the record)
		if !sync.isSkipped() {
			m.syncTransaction = sync
		}
	}

	// If we are external and the user disabled incoming transaction checking, check outputs
	if m.isExternal() && !m.Client().IsITCEnabled() {
		// Check that the transaction has >= 1 known destination
		if !m.TransactionBase.hasOneKnownDestination(ctx, m.Client(), m.GetOptions(false)...) {
			return ErrNoMatchingOutputs
		}
	}

	// Process the UTXOs
	if err = m.processUtxos(ctx); err != nil {
		return err
	}

	// Set the values from the inputs/outputs and draft tx
	m.TotalValue, m.Fee = m.getValues()

	// Add values if found
	if m.TransactionBase.parsedTx != nil {
		m.NumberOfInputs = uint32(len(m.TransactionBase.parsedTx.Inputs))
		m.NumberOfOutputs = uint32(len(m.TransactionBase.parsedTx.Outputs))
	}

	m.DebugLog("end: " + m.Name() + " BeforeCreating hook")
	m.beforeCreateCalled = true
	return nil
}

// AfterCreated will fire after the model is created in the Datastore
func (m *Transaction) AfterCreated(ctx context.Context) error {
	m.DebugLog("starting: " + m.Name() + " AfterCreated hook...")

	// Pre-build the options
	opts := m.GetOptions(false)

	// update the xpub balances
	for xPubID, balance := range m.XpubOutputValue {
		// todo: run this in a go routine? (move this into a function on the xpub model?)
		xPub, err := getXpubWithCache(ctx, m.Client(), "", xPubID, opts...)
		if err != nil {
			return err
		} else if xPub == nil {
			return ErrMissingRequiredXpub
		}
		if err = xPub.incrementBalance(ctx, balance); err != nil {
			return err
		}
	}

	// Update the draft transaction, process broadcasting
	// todo: go routine (however it's not working, panic in save for missing datastore)
	if m.draftTransaction != nil {
		m.draftTransaction.Status = DraftStatusComplete
		m.draftTransaction.FinalTxID = m.ID
		if err := m.draftTransaction.Save(ctx); err != nil {
			return err
		}
	}

	// Fire notifications (this is already in a go routine)
	notify(notifications.EventTypeCreate, m)

	m.DebugLog("end: " + m.Name() + " AfterCreated hook")
	return nil
}

// AfterUpdated will fire after the model is updated in the Datastore
func (m *Transaction) AfterUpdated(_ context.Context) error {
	m.DebugLog("starting: " + m.Name() + " AfterUpdated hook...")

	// Fire notifications (this is already in a go routine)
	notify(notifications.EventTypeUpdate, m)

	m.DebugLog("end: " + m.Name() + " AfterUpdated hook")
	return nil
}

// AfterDeleted will fire after the model is deleted in the Datastore
func (m *Transaction) AfterDeleted(_ context.Context) error {
	m.DebugLog("starting: " + m.Name() + " AfterDelete hook...")

	// Fire notifications (this is already in a go routine)
	notify(notifications.EventTypeDelete, m)

	m.DebugLog("end: " + m.Name() + " AfterDelete hook")
	return nil
}

// ChildModels will get any related sub models
func (m *Transaction) ChildModels() (childModels []ModelInterface) {
	// Add the UTXOs if found
	for index := range m.utxos {
		childModels = append(childModels, &m.utxos[index])
	}

	// Add the broadcast transaction record
	if m.syncTransaction != nil {
		childModels = append(childModels, m.syncTransaction)
	}

	return
}

// Migrate model specific migration on startup
func (m *Transaction) Migrate(client datastore.ClientInterface) error {
	tableName := client.GetTableName(tableTransactions)
	if client.Engine() == datastore.MySQL {
		if err := m.migrateMySQL(client, tableName); err != nil {
			return err
		}
	} else if client.Engine() == datastore.PostgreSQL {
		if err := m.migratePostgreSQL(client, tableName); err != nil {
			return err
		}
	}

	err := m.migrateBUMP()
	if err != nil {
		return err
	}

	return client.IndexMetadata(tableName, xPubMetadataField)
}

// migratePostgreSQL is specific migration SQL for Postgresql
func (m *Transaction) migratePostgreSQL(client datastore.ClientInterface, tableName string) error {
	tx := client.Execute(`CREATE INDEX IF NOT EXISTS idx_` + tableName + `_xpub_in_ids ON ` +
		tableName + ` USING gin (xpub_in_ids jsonb_ops)`)
	if tx.Error != nil {
		return tx.Error
	}

	if tx = client.Execute(`CREATE INDEX IF NOT EXISTS idx_` + tableName + `_xpub_out_ids ON ` +
		tableName + ` USING gin (xpub_out_ids jsonb_ops)`); tx.Error != nil {
		return tx.Error
	}

	return nil
}

// migrateMySQL is specific migration SQL for MySQL
func (m *Transaction) migrateMySQL(client datastore.ClientInterface, tableName string) error {
	idxName := "idx_" + tableName + "_xpub_in_ids"
	idxExists, err := client.IndexExists(tableName, idxName)
	if err != nil {
		return err
	}
	if !idxExists {
		tx := client.Execute("ALTER TABLE `" + tableName + "`" +
			" ADD INDEX " + idxName + " ( (CAST(xpub_in_ids AS CHAR(64) ARRAY)) )")
		if tx.Error != nil {
			m.Client().Logger().Error(context.Background(), "failed creating json index on mysql: "+tx.Error.Error())
			return nil //nolint:nolintlint,nilerr // error is not needed
		}
	}

	idxName = "idx_" + tableName + "_xpub_out_ids"
	if idxExists, err = client.IndexExists(
		tableName, idxName,
	); err != nil {
		return err
	}
	if !idxExists {
		tx := client.Execute("ALTER TABLE `" + tableName + "`" +
			" ADD INDEX " + idxName + " ( (CAST(xpub_out_ids AS CHAR(64) ARRAY)) )")
		if tx.Error != nil {
			m.Client().Logger().Error(context.Background(), "failed creating json index on mysql: "+tx.Error.Error())
			return nil //nolint:nolintlint,nilerr // error is not needed
		}
	}

	tx := client.Execute("ALTER TABLE `" + tableName + "` MODIFY COLUMN hex longtext")
	if tx.Error != nil {
		m.Client().Logger().Error(context.Background(), "failed changing hex type to longtext in MySQL: "+tx.Error.Error())
		return nil //nolint:nolintlint,nilerr // error is not needed
	}

	return nil
}

func (m *Transaction) migrateBUMP() error {
	ctx := context.Background()
	txs, err := getTransactionsToCalculateBUMP(ctx, nil, WithClient(m.client))
	if err != nil {
		return err
	}
	for _, tx := range txs {
		bump := tx.MerkleProof.ToBUMP()
		bump.BlockHeight = tx.BlockHeight
		tx.BUMP = bump
		_ = tx.Save(ctx)
	}
	return nil
}
