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

func (m *Metadata) AddOrUpdateBlockMeta(fileName string, blockMeta BlockMeta) error {
	blockInfo, err := m.GetBlockInfo(fileName)
	if err != nil {
		return err
	}
	blockInfo[blockMeta.BlockID] = blockMeta
	return nil
}

func (m *Metadata) DelFile(fileName string) {
	delete(m.FileInfo, fileName)
}
