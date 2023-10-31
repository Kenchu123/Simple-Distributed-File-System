package metadata

import "fmt"

// Metadata handle metadata.
type Metadata struct {
	FileInfo map[string]BlockInfo // map[fileName]BlockInfo}
}

// BlockInfo is the metadata of a file.
type BlockInfo map[int64]BlockMeta // map[blockID]BlockMeta}

// BlockMeta is the metadata of a file block.
type BlockMeta struct {
	HostNames []string
	FileName  string
	BlockID   int64
}

// NewMetadata creates a new metadata.
func NewMetadata() *Metadata {
	return &Metadata{
		FileInfo: make(map[string]BlockInfo),
	}
}

func (m *Metadata) IsFileExist(fileName string) bool {
	_, ok := m.FileInfo[fileName]
	return ok
}

// AddFileBlock adds a file block to metadata.
func (m *Metadata) AddFileBlock(hostname, fileName string, blockID int64) {
	if _, ok := m.FileInfo[fileName]; !ok {
		m.FileInfo[fileName] = make(map[int64]BlockMeta)
	}
	m.FileInfo[fileName][int64(blockID)] = BlockMeta{
		HostNames: append(m.FileInfo[fileName][int64(blockID)].HostNames, hostname),
		FileName:  fileName,
		BlockID:   blockID,
	}
}

// AddOrUpdateFile adds or updates a file to metadata.
func (m *Metadata) AddOrUpdateFile(fileName string, blockInfo BlockInfo) {
	m.FileInfo[fileName] = blockInfo
}

func (m *Metadata) GetBlockInfo(fileName string) (BlockInfo, error) {
	if _, ok := m.FileInfo[fileName]; !ok {
		return nil, fmt.Errorf("file %s not found", fileName)
	}
	return m.FileInfo[fileName], nil
}

func (m *Metadata) GetBlockMeta(fileName string, blockID int64) (BlockMeta, error) {
	blockInfo, err := m.GetBlockInfo(fileName)
	if err != nil {
		return BlockMeta{}, err
	}
	if _, ok := blockInfo[blockID]; !ok {
		return BlockMeta{}, fmt.Errorf("block %d of file %s not found", blockID, fileName)
	}
	return blockInfo[blockID], nil
}
