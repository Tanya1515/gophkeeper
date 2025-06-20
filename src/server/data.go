package server

import (
	"crypto/cipher"
	"sync"

	"go.uber.org/zap"

	dataStorage "github.com/Tanya1515/gophkeeper.git/src/data_storage"
	fileStorage "github.com/Tanya1515/gophkeeper.git/src/file_storage"
	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
)

// Crypto - structure, that containts object for generating
// key for encryption/decryption data.
type Crypto struct {
	Aesgcm cipher.AEAD
}

// GophkeeperServer - structure, that containts data about Gophkeeper application
type GophkeeperServer struct {
	DataStorage dataStorage.DataStorage // DataStorage saves all user sensetive data

	FileStorage fileStorage.FileStorage // FileStorage saves all user files

	Logger zap.SugaredLogger // Logger saves all server info

	UserOTP map[string]string // UserOTP saves all one-time passwords for users

	Mutex *sync.Mutex // Mutex for synchronization

	Crypto // Crypto data for encryption/decryption sensetive information

	pb.UnimplementedGophkeeperServer // type pb.Unimplemented<TypeName> is used for backward compatibility
}
