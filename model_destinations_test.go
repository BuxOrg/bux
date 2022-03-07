package bux

import (
	"testing"

	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/tester"
	"github.com/BuxOrg/bux/utils"
	"github.com/DATA-DOG/go-sqlmock"
	bscript2 "github.com/libsv/go-bt/v2/bscript"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// todo: finish unit tests!

var testLockingScript = "76a9147ff514e6ae3deb46e6644caac5cdd0bf2388906588ac"
var testAddressID = "fc1e635d98151c6008f29908ee2928c60c745266f9853e945c917b1baa05973e"
var testDestinationID = "c775e7b757ede630cd0aa1113bd102661ab38829ca52a6422ab782862f268646"

// TestDestination_newDestination will test the method newDestination()
func TestDestination_newDestination(t *testing.T) {
	t.Parallel()

	t.Run("New empty destination model", func(t *testing.T) {
		destination := newDestination("", "", New())
		require.NotNil(t, destination)
		assert.IsType(t, Destination{}, *destination)
		assert.Equal(t, ModelDestination.String(), destination.GetModelName())
		assert.Equal(t, true, destination.IsNew())
		assert.Equal(t, "", destination.LockingScript)
		assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", destination.GetID())
	})

	t.Run("New destination model", func(t *testing.T) {
		testScript := "1234567890"
		xPubID := "xpub123456789"
		destination := newDestination(xPubID, testScript, New())
		require.NotNil(t, destination)
		assert.IsType(t, Destination{}, *destination)
		assert.Equal(t, ModelDestination.String(), destination.GetModelName())
		assert.Equal(t, true, destination.IsNew())
		assert.Equal(t, testScript, destination.LockingScript)
		assert.Equal(t, xPubID, destination.XpubID)
		assert.Equal(t, bscript2.ScriptTypeNonStandard, destination.Type)
		assert.Equal(t, testDestinationID, destination.GetID())
	})
}

// TestDestination_newAddress will test the method newAddress()
func TestDestination_newAddress(t *testing.T) {
	t.Parallel()

	t.Run("New empty address model", func(t *testing.T) {
		address, err := newAddress("", 0, 0, New())
		assert.Nil(t, address)
		assert.Error(t, err)
	})

	t.Run("invalid xPub", func(t *testing.T) {
		address, err := newAddress("test", 0, 0, New())
		assert.Nil(t, address)
		assert.Error(t, err)
	})

	t.Run("valid xPub", func(t *testing.T) {
		address, err := newAddress(testXPub, 0, 0, New())
		require.NotNil(t, address)
		require.NoError(t, err)

		// Default values
		assert.IsType(t, Destination{}, *address)
		assert.Equal(t, ModelDestination.String(), address.GetModelName())
		assert.Equal(t, true, address.IsNew())

		// Check set address
		assert.Equal(t, testXPubID, address.XpubID)
		assert.Equal(t, testExternalAddress, address.Address)

		// Check set locking script
		assert.Equal(t, testLockingScript, address.LockingScript)
		assert.Equal(t, bscript2.ScriptTypePubKeyHash, address.Type)
		assert.Equal(t, testAddressID, address.GetID())
	})

}

// TestDestination_GetModelName will test the method GetModelName()
func TestDestination_GetModelName(t *testing.T) {
	t.Parallel()

	t.Run("model name", func(t *testing.T) {
		address, err := newAddress(testXPub, 0, 0, New())
		require.NotNil(t, address)
		require.NoError(t, err)

		assert.Equal(t, ModelDestination.String(), address.GetModelName())
	})
}

// TestDestination_GetID will test the method GetID()
func TestDestination_GetID(t *testing.T) {
	t.Parallel()

	t.Run("valid id - address", func(t *testing.T) {
		address, err := newAddress(testXPub, 0, 0, New())
		require.NotNil(t, address)
		require.NoError(t, err)

		assert.Equal(t, testAddressID, address.GetID())
	})

	t.Run("valid id - destination", func(t *testing.T) {
		testScript := "1234567890"
		xPubID := "xpub123456789"
		destination := newDestination(xPubID, testScript, New())
		require.NotNil(t, destination)

		assert.Equal(t, testDestinationID, destination.GetID())
	})
}

// TestDestination_BeforeCreating will test the method BeforeCreating()
func TestDestination_BeforeCreating(t *testing.T) {
	// finish test
}

// TestDestination_Save will test the method Save()
func TestDestination_Save(t *testing.T) {
	// finish test
}

// TestDestination_setLockingScriptForAddress will test the method setLockingScriptForAddress()
func TestDestination_setLockingScriptForAddress(t *testing.T) {
	// finish test
}

// TestDestination_setAddress will test the method setAddress()
func TestDestination_setAddress(t *testing.T) {
	// finish test
}

// TestDestination_getDestinationByID will test the method getDestinationByID()
func TestDestination_getDestinationByID(t *testing.T) {

	t.Run("does not exist", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()

		xPub, err := getDestinationByID(ctx, testDestinationID, client.DefaultModelOptions()...)
		assert.NoError(t, err)
		assert.Nil(t, xPub)
	})

	t.Run("get", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()

		destination := newDestination(testXPubID, testLockingScript, client.DefaultModelOptions()...)
		err := destination.Save(ctx)
		assert.NoError(t, err)

		gDestination, gErr := getDestinationByID(ctx, destination.ID, client.DefaultModelOptions()...)
		assert.NoError(t, gErr)
		assert.IsType(t, Destination{}, *gDestination)
		assert.Equal(t, testXPubID, gDestination.XpubID)
		assert.Equal(t, testLockingScript, gDestination.LockingScript)
		assert.Equal(t, testExternalAddress, gDestination.Address)
		assert.Equal(t, utils.ScriptTypePubKeyHash, gDestination.Type)
		assert.Equal(t, uint32(0), gDestination.Chain)
		assert.Equal(t, uint32(0), gDestination.Num)
	})
}

// TestDestination_getDestinationByAddress will test the method getDestinationByAddress()
func TestDestination_getDestinationByAddress(t *testing.T) {

	t.Run("does not exist", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()

		xPub, err := getDestinationByAddress(ctx, testExternalAddress, client.DefaultModelOptions()...)
		assert.NoError(t, err)
		assert.Nil(t, xPub)
	})

	t.Run("get", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()

		destination := newDestination(testXPubID, testLockingScript, client.DefaultModelOptions()...)
		err := destination.Save(ctx)
		assert.NoError(t, err)

		gDestination, gErr := getDestinationByAddress(ctx, testExternalAddress, client.DefaultModelOptions()...)
		assert.NoError(t, gErr)
		assert.IsType(t, Destination{}, *gDestination)
		assert.Equal(t, testXPubID, gDestination.XpubID)
		assert.Equal(t, testLockingScript, gDestination.LockingScript)
		assert.Equal(t, testExternalAddress, gDestination.Address)
		assert.Equal(t, utils.ScriptTypePubKeyHash, gDestination.Type)
		assert.Equal(t, uint32(0), gDestination.Chain)
		assert.Equal(t, uint32(0), gDestination.Num)
	})
}

// TestDestination_getDestinationByLockingScript will test the method getDestinationByLockingScript()
func TestDestination_getDestinationByLockingScript(t *testing.T) {

	t.Run("does not exist", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()

		xPub, err := getDestinationByLockingScript(ctx, testLockingScript, client.DefaultModelOptions()...)
		assert.NoError(t, err)
		assert.Nil(t, xPub)
	})

	t.Run("get destination", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()

		destination := newDestination(testXPubID, testLockingScript, client.DefaultModelOptions()...)
		err := destination.Save(ctx)
		assert.NoError(t, err)

		gDestination, gErr := getDestinationByLockingScript(ctx, testLockingScript, client.DefaultModelOptions()...)
		assert.NoError(t, gErr)
		assert.IsType(t, Destination{}, *gDestination)
		assert.Equal(t, testXPubID, gDestination.XpubID)
		assert.Equal(t, testLockingScript, gDestination.LockingScript)
		assert.Equal(t, testExternalAddress, gDestination.Address)
		assert.Equal(t, utils.ScriptTypePubKeyHash, gDestination.Type)
		assert.Equal(t, uint32(0), gDestination.Chain)
		assert.Equal(t, uint32(0), gDestination.Num)
	})
}

// BenchmarkDestination_newAddress will test the method newAddress()
func BenchmarkDestination_newAddress(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = newAddress(testXPub, 0, 0, New())
	}
}

// TestClient_NewDestination will test the method NewDestination()
func TestClient_NewDestination(t *testing.T) {

	t.Run("valid - simple destination", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, true, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()

		// Get new random key
		_, xPub, rawXPub := CreateNewXPub(ctx, t, client)
		require.NotNil(t, xPub)

		opts := append(
			client.DefaultModelOptions(),
			WithMetadatas(map[string]interface{}{
				ReferenceIDField: "some-reference-id",
				testMetadataKey:  testMetadataValue,
			}),
		)

		// Create a new destination
		destination, err := client.NewDestination(
			ctx, rawXPub, utils.ChainExternal, utils.ScriptTypePubKeyHash, opts...,
		)
		require.NoError(t, err)
		require.NotNil(t, destination)
		assert.Equal(t, "some-reference-id", destination.Metadata[ReferenceIDField])
		assert.Equal(t, 64, len(destination.ID))
		assert.Greater(t, len(destination.Address), 32)
		assert.Greater(t, len(destination.LockingScript), 32)
		assert.Equal(t, utils.ScriptTypePubKeyHash, destination.Type)
		assert.Equal(t, uint32(0), destination.Num)
		assert.Equal(t, utils.ChainExternal, destination.Chain)
		assert.Equal(t, Metadata{ReferenceIDField: "some-reference-id", testMetadataKey: testMetadataValue}, destination.Metadata)
		assert.Equal(t, xPub.ID, destination.XpubID)
	})

	t.Run("error - invalid xPub", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, true, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()

		opts := append(
			client.DefaultModelOptions(),
			WithMetadatas(map[string]interface{}{
				testMetadataKey: testMetadataValue,
			}),
		)

		// Create a new destination
		destination, err := client.NewDestination(
			ctx, "bad-value", utils.ChainExternal, utils.ScriptTypePubKeyHash,
			opts...,
		)
		require.Error(t, err)
		require.Nil(t, destination)
	})

	t.Run("error - xPub not found", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, true, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()

		opts := append(
			client.DefaultModelOptions(),
			WithMetadatas(map[string]interface{}{
				testMetadataKey: testMetadataValue,
			}),
		)

		// Create a new destination
		destination, err := client.NewDestination(
			ctx, testXPub, utils.ChainExternal, utils.ScriptTypePubKeyHash,
			opts...,
		)
		require.Error(t, err)
		require.Nil(t, destination)
		assert.ErrorIs(t, err, ErrMissingXpub)
	})

	t.Run("error - unsupported destination type", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, true, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()

		// Get new random key
		_, xPub, rawXPub := CreateNewXPub(ctx, t, client)
		require.NotNil(t, xPub)

		opts := append(
			client.DefaultModelOptions(),
			WithMetadatas(map[string]interface{}{
				testMetadataKey: testMetadataValue,
			}),
		)

		// Create a new destination
		destination, err := client.NewDestination(
			ctx, rawXPub, utils.ChainExternal, utils.ScriptTypeMultiSig,
			opts...,
		)
		require.Error(t, err)
		require.Nil(t, destination)
	})
}

// TestDestination_Save will test the method Save()
func (ts *EmbeddedDBTestSuite) TestDestination_Save() {
	ts.T().Run("[sqlite] [redis] [mocking] - create destination", func(t *testing.T) {
		tc := ts.genericMockedDBClient(t, datastore.SQLite)
		defer tc.Close(tc.ctx)

		xPub := newXpub(testXPub, append(tc.client.DefaultModelOptions(), New())...)
		require.NotNil(t, xPub)

		destination := newDestination(xPub.ID, testLockingScript, append(tc.client.DefaultModelOptions(), New())...)
		require.NotNil(t, destination)
		destination.DraftID = testDraftID

		// Create the expectations
		tc.MockSQLDB.ExpectBegin()

		// Create model
		tc.MockSQLDB.ExpectExec("INSERT INTO `"+tc.tablePrefix+"_destinations` ("+
			"`created_at`,`updated_at`,`metadata`,`deleted_at`,`id`,`xpub_id`,`locking_script`,"+
			"`type`,`chain`,`num`,`address`,`draft_id`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)").WithArgs(
			tester.AnyTime{},    // created_at
			tester.AnyTime{},    // updated_at
			nil,                 // metadata
			nil,                 // deleted_at
			tester.AnyGUID{},    // id
			xPub.GetID(),        // xpub_id
			testLockingScript,   // locking_script
			destination.Type,    // type
			0,                   // chain
			0,                   // num
			destination.Address, // address
			testDraftID,         // draft_id
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Commit the TX
		tc.MockSQLDB.ExpectCommit()

		err := destination.Save(tc.ctx)
		require.NoError(t, err)

		err = tc.MockSQLDB.ExpectationsWereMet()
		require.NoError(t, err)
	})

	ts.T().Run("[mysql] [redis] [mocking] - create destination", func(t *testing.T) {
		tc := ts.genericMockedDBClient(t, datastore.MySQL)
		defer tc.Close(tc.ctx)

		xPub := newXpub(testXPub, append(tc.client.DefaultModelOptions(), New())...)
		require.NotNil(t, xPub)

		destination := newDestination(xPub.ID, testLockingScript, append(tc.client.DefaultModelOptions(), New())...)
		require.NotNil(t, destination)
		destination.DraftID = testDraftID

		// Create the expectations
		tc.MockSQLDB.ExpectBegin()

		// Create model
		tc.MockSQLDB.ExpectExec("INSERT INTO `"+tc.tablePrefix+"_destinations` ("+
			"`created_at`,`updated_at`,`metadata`,`deleted_at`,`id`,`xpub_id`,`locking_script`,"+
			"`type`,`chain`,`num`,`address`,`draft_id`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)").WithArgs(
			tester.AnyTime{},    // created_at
			tester.AnyTime{},    // updated_at
			nil,                 // metadata
			nil,                 // deleted_at
			tester.AnyGUID{},    // id
			xPub.GetID(),        // xpub_id
			testLockingScript,   // locking_script
			destination.Type,    // type
			0,                   // chain
			0,                   // num
			destination.Address, // address
			testDraftID,         // draft_id
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Commit the TX
		tc.MockSQLDB.ExpectCommit()

		err := destination.Save(tc.ctx)
		require.NoError(t, err)

		err = tc.MockSQLDB.ExpectationsWereMet()
		require.NoError(t, err)
	})

	ts.T().Run("[postgresql] [redis] [mocking] - create destination", func(t *testing.T) {
		tc := ts.genericMockedDBClient(t, datastore.PostgreSQL)
		defer tc.Close(tc.ctx)

		xPub := newXpub(testXPub, append(tc.client.DefaultModelOptions(), New())...)
		require.NotNil(t, xPub)

		destination := newDestination(xPub.ID, testLockingScript, append(tc.client.DefaultModelOptions(), New())...)
		require.NotNil(t, destination)
		destination.DraftID = testDraftID

		// Create the expectations
		tc.MockSQLDB.ExpectBegin()

		// Create model
		tc.MockSQLDB.ExpectExec(`INSERT INTO "`+tc.tablePrefix+`_destinations" ("created_at","updated_at","metadata","deleted_at","id","xpub_id","locking_script","type","chain","num","address","draft_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`).WithArgs(
			tester.AnyTime{},    // created_at
			tester.AnyTime{},    // updated_at
			nil,                 // metadata
			nil,                 // deleted_at
			tester.AnyGUID{},    // id
			xPub.GetID(),        // xpub_id
			testLockingScript,   // locking_script
			destination.Type,    // type
			0,                   // chain
			0,                   // num
			destination.Address, // address
			testDraftID,         // draft_id
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Commit the TX
		tc.MockSQLDB.ExpectCommit()

		err := destination.Save(tc.ctx)
		require.NoError(t, err)

		err = tc.MockSQLDB.ExpectationsWereMet()
		require.NoError(t, err)
	})

	ts.T().Run("[mongo] [redis] [mocking] - create destination", func(t *testing.T) {
		// todo: mocking for MongoDB
	})
}
