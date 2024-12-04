package util

type UploadConfig struct {
	File, Protocol, SIP, SPort, CIP, CPort string
	ChunkSize                              int32
}
