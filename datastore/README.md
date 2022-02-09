# Quickstart on Datastore's

### PostgreSQL 
```sql
CREATE DATABASE bux;
CREATE USER buxdblocaluser WITH ENCRYPTED PASSWORD '_LocalSecretPassForDb456_';
GRANT ALL PRIVILEGES ON DATABASE bux TO buxdblocaluser;
ALTER ROLE buxdblocaluser SUPERUSER CREATEDB CREATEROLE INHERIT LOGIN;
GRANT CREATE ON SCHEMA public TO buxdblocaluser;
GRANT ALL ON SCHEMA public TO buxdblocaluser;
```

### SQLite
```text
Set a database path (IE: default_db.db)
```

### MySQL
```sql
CREATE DATABASE bux;
DROP USER IF EXISTS 'buxDbLocalUser'@'%';
SET GLOBAL sql_mode = '';
SET GLOBAL sql_mode = 'STRICT_TRANS_TABLES,ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION';
CREATE USER 'buxDbLocalUser'@'%' IDENTIFIED BY '_LocalSecretPassForDb789_';
GRANT USAGE ON `bux`.* TO 'buxDbLocalUser'@'%';
GRANT DELETE ON `bux`.* TO 'buxDbLocalUser'@'%';
GRANT INSERT ON `bux`.* TO 'buxDbLocalUser'@'%';
GRANT SELECT ON `bux`.* TO 'buxDbLocalUser'@'%';
GRANT UPDATE ON `bux`.* TO 'buxDbLocalUser'@'%';
GRANT CREATE ON `bux`.* TO 'buxDbLocalUser'@'%';
GRANT INDEX ON `bux`.* TO 'buxDbLocalUser'@'%';
GRANT ALTER ON `bux`.* TO 'buxDbLocalUser'@'%';
GRANT REFERENCES ON `bux`.* TO 'buxDbLocalUser'@'%';
GRANT DROP ON `bux`.* TO 'buxDbLocalUser'@'%';
GRANT LOCK TABLES ON `bux`.* TO 'buxDbLocalUser'@'%';
GRANT CREATE TEMPORARY TABLES ON `bux`.* TO 'buxDbLocalUser'@'%';
GRANT TRIGGER ON `bux`.* TO 'buxDbLocalUser'@'%';
GRANT SUPER ON *.* TO 'buxDbLocalUser'@'%';
FLUSH PRIVILEGES;
```

### MongoDB
```text
No setup required
```