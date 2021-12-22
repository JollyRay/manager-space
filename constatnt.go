package main

/* Rules */
const (
	LOGIN   byte = iota
	LOGOUT  byte = iota
	DATA    byte = iota
	EXIT    byte = iota //TODO: WHERE? Nobody set this param
	COMMAND byte = iota
)

/* Reply */
const (
	REFUSAL          byte = 1<<7 + iota
	ACCEPT           byte = 1<<7 + iota
	CONFIRMATION     byte = 1<<7 + iota
	RUNTIMEERROR     byte = 1<<7 + iota
	PERMISSIONDENIED byte = 1<<7 + iota
	CARIMG           byte = 1<<7 + iota
)

/* Command */
const (
	SHOWPILOTE      byte = iota
	SHOWCAR         byte = iota
	SHOWCLIENT      byte = iota
	SHOWQUEUE       byte = iota
	SHOWRULE        byte = iota
	ADDCAR          byte = iota
	ADDCLIENT       byte = iota
	ADDMANAGER      byte = iota
	DELETECAR       byte = iota
	DELETECLIENT    byte = iota
	CLOSECONNECTION byte = iota
	PREPARECAR      byte = iota
	PREPARECLIENT   byte = iota
	BANCAR          byte = iota
	BANCLIENT       byte = iota
	SETRULE         byte = iota
)

/* Command Rule */

const (
	TIMEDELEY_NUM           byte = iota
	TIMEOUT_NUM             byte = iota
	USERBAN_NUM             byte = iota
	CARBAN_NUM              byte = iota
	FORCEDISCONNECTUSER_NUM byte = iota
	FORCEDISCONNECTCAR_NUM  byte = iota
)

const (
	_                        = 1 << iota
	_                        = 1 << iota
	_                        = 1 << iota
	_                        = 1 << iota
	USERBAN             byte = 1 << iota
	CARBAN              byte = 1 << iota
	FORCEDISCONNECTUSER byte = 1 << iota
	FORCEDISCONNECTCAR  byte = 1 << iota
)
