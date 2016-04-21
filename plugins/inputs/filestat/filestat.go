package filestat

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
  ## Files to gather stats about.
  files = [""]
  ## If true, read the entire file and calculate an md5 checksum.
  md5 = false
`

type FileStat struct {
	Md5   bool
	Files []string
}

func (_ *FileStat) Description() string {
	return "Read stats about given file(s)"
}

func (_ *FileStat) SampleConfig() string { return sampleConfig }

func (f *FileStat) Gather(acc telegraf.Accumulator) error {
	var errS string
	for _, file := range f.Files {
		// allocate default tags and fields
		tags := map[string]string{
			"file": file,
		}
		fields := map[string]interface{}{
			"exists": int64(0),
		}

		// Get file stats
		fileInfo, err := os.Stat(file)
		if os.IsNotExist(err) {
			// file doesn't exist, so move on to the next
			acc.AddFields("filestat", fields, tags)
			continue
		}
		if err != nil {
			errS += err.Error() + " "
			continue
		}

		// file exists and no errors encountered
		fields["exists"] = int64(1)
		fields["size_bytes"] = fileInfo.Size()
		fields["mode"] = fileInfo.Mode().String()

		if f.Md5 {
			of, err := os.Open(file)
			if err != nil {
				errS += err.Error() + " "
			} else {
				defer of.Close()

				hash := md5.New()
				_, err = io.Copy(hash, of)
				if err != nil {
					// fatal error
					return err
				}
				fields["md5_sum"] = fmt.Sprintf("%x", hash.Sum(nil))
			}
		}

		acc.AddFields("filestat", fields, tags)
	}

	if errS != "" {
		return fmt.Errorf(errS)
	}
	return nil
}

func init() {
	inputs.Add("filestat", func() telegraf.Input {
		return &FileStat{}
	})
}
