/******************************************************************************
 * 5G-MAG Reference Tools: CMMF Encoder
 ******************************************************************************
 * Copyright: (C)2024-2026 Dolby Laboratories Inc.
 * License: 5G-MAG Public License v1
 *
 * Licensed under the License terms and conditions for use, reproduction, and
 * distribution of 5G-MAG software (the “License”).  You may not use this file
 * except in compliance with the License.  You may obtain a copy of the License at
 * https://www.5g-mag.com/reference-tools.  Unless required by applicable law or
 * agreed to in writing, software distributed under the License is distributed on
 * an “AS IS” BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied.
 *
 * See the License for the specific language governing permissions and limitations
 * under the License.
 */

package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"gitlab-sfo.dolby.net/interferex/cmmf_encoder/cmmf_encoder"
)

// isValidFile checks whether a given file is valid: i.e. exists, and is not a directory
func isValidFile(filename string) (bool, error) {
	fStat, err := os.Stat(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, fmt.Errorf("file %s does not exist", filename)
		}
		return false, fmt.Errorf("cannot read file %s: %w", filename, err)
	}
	if fStat.IsDir() {
		return false, fmt.Errorf("%s is a directory", filename)
	}
	return true, nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	pathEncoderConfigFile := flag.String("enc-config", "", "path to the encoder config file")
	pathSourceContentFile := flag.String("src-content", "", "path to the source content file")
	pathEncodedOutputFile := flag.String("enc-output", "", "path to the encoded output file")
	flag.Parse()

	if *pathEncoderConfigFile == "" || *pathSourceContentFile == "" || *pathEncodedOutputFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	// in this example we shall encode 16384 bytes of source content into 8 packets
	srcContentFileIsValid, err := isValidFile(*pathSourceContentFile)
	if err != nil || !srcContentFileIsValid {
		logger.Error("source content file is not valid", "file", *pathSourceContentFile, "error", err)
		os.Exit(1)
	}

	sourceBytes := make([]byte, 16384)
	f, err := os.Open(*pathSourceContentFile)
	if err != nil {
		logger.Error("failed to open source content file", "file", *pathSourceContentFile, "error", err)
		os.Exit(1)
	}
	defer f.Close()

	n, err := io.ReadFull(f, sourceBytes)
	if err != nil {
		logger.Error("failed to read source content file", "file", *pathSourceContentFile, "error", err)
		os.Exit(1)
	}
	logger.Info("read bytes from source", "n", n)

	sourceBytesReader := bytes.NewReader(sourceBytes)

	// read in the encoder configuration parameters from the JSON file
	encConfFileIsValid, err := isValidFile(*pathEncoderConfigFile)
	if err != nil || !encConfFileIsValid {
		logger.Error("encoding config file is not valid", "file", *pathEncoderConfigFile, "error", err)
		os.Exit(1)
	}

	fileBytes, err := os.ReadFile(*pathEncoderConfigFile)
	if err != nil {
		logger.Error("error reading config file", "file", *pathEncoderConfigFile, "error", err)
		os.Exit(1)
	}

	encConfigFileReader := bytes.NewReader(fileBytes)
	enc, err := cmmf_encoder.NewETSIEncoder(sourceBytesReader, encConfigFileReader)
	if err != nil {
		logger.Error("error creating encoder", "error", err)
		os.Exit(1)
	}

	// write the encoded bitstream
	var outBitstream bytes.Buffer
	outBitstreamWriter := bufio.NewWriter(&outBitstream)
	err = enc.EncodeVariant("cmmf-a", outBitstreamWriter)
	if err != nil {
		logger.Error("error encoding the 'cmmf-a' variant", "error", err)
		os.Exit(1)
	}
	err = outBitstreamWriter.Flush()
	if err != nil {
		logger.Error("error while attempting to flush the buffered writer", "error", err)
		os.Exit(1)
	}

	err = os.WriteFile(*pathEncodedOutputFile, outBitstream.Bytes(), 0666)
	if err != nil {
		logger.Error("failed to write encoded bitstream file", "file", *pathEncodedOutputFile, "error", err)
		os.Exit(1)
	}
}
