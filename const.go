package qbchain

const (
	BLOCKCHAIN_PORT      = "9119"
	MAX_NODE_CONNECTIONS = 400

	NETWORK_KEY_SIZE = 80

	TRANSACTION_HEADER_SIZE = NETWORK_KEY_SIZE /* from key */ + NETWORK_KEY_SIZE /* to key */ + 4 /* int32 timestamp */ + 32 /* sha256 payload hash */ + 4 /* int32 payload length */ + 4 /* int32 nonce */
	BLOCK_HEADER_SIZE       = NETWORK_KEY_SIZE /* origin key */ + 4 /* int32 timestamp */ + 32 /* prev block hash */ + 32 /* merkel tree hash */ + 4                                      /* int32 nonce */

	KEY_POW_COMPLEXITY = 0

	TRANSACTION_POW_COMPLEXITY = 1

	BLOCK_POW_COMPLEXITY = 2

	KEY_SIZE = 28

	POW_PREFIX = 0

	MESSAGE_TYPE_SIZE    = 1
	MESSAGE_OPTIONS_SIZE = 4

	DB_NAMESPACE = "qbchain"
)
