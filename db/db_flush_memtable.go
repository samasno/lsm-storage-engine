package db

var DatafileBlocksize uint64 = 4096

func (db *DB) FlushMemtableToDisk() error {
	df, err := db.manifest.NewDatafile()
	if err != nil {
		return err
	}

	block := NewDataBlock()
	scanner := db.memtable.Scanner(nil)

	offset := uint64(0)
	for scanner.Next() {
		entry := NewDatablockEntry(scanner.Key(), scanner.Value())
		appended, err := block.Append(entry)
		if err != nil {
			return err
		}

		if !appended {
			_, err = df.Write(block.Bytes())
			if err != nil {
				return err
			}

			blockIndex := NewDatablockIndexEntry(block.LastKey(), offset)
			df.Index.Append(blockIndex)
			offset += DatafileBlocksize

			block = NewDataBlock()
			block.Append(entry)
		}
	}

	_, err = df.Write(block.Bytes())
	if err != nil {
		return err
	}

	blockIndex := NewDatablockIndexEntry(block.LastKey(), offset)
	df.Index.Append(blockIndex)

	offset += DatafileBlocksize
	indexB, err := df.Index.Bytes()
	if err != nil {
		return err
	}

	_, err = df.Write(indexB)
	if err != nil {
		return err
	}

	footer, err := GenerateFooter(uint64(len(indexB)), offset)
	if err != nil {
		return err
	}

	_, err = df.Write(footer)
	if err != nil {
		return err
	}

	return nil
}
