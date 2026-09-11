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

package cmmf_encoder

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type ETSIConfigError struct {
	arg     string
	message string
}

func (e ETSIConfigError) Error() string {
	return fmt.Sprintf("ETSI encoder config parsing error for field %s: %s", e.arg, e.message)
}

// Set is a minimal implementation of a set of unique generic elements.
// It was primarily written to check (where applicable) whether a given value in the config file is one of an expected
// set of values.
type Set[T comparable] struct {
	Members map[T]struct{}
}

func NewSet[T comparable](elems []T) *Set[T] {
	newSet := &Set[T]{
		Members: make(map[T]struct{}, len(elems)),
	}
	for _, elem := range elems {
		newSet.Members[elem] = struct{}{}
	}
	return newSet
}

func (s *Set[T]) Contains(elem T) bool {
	if _, ok := s.Members[elem]; ok {
		return true
	}
	return false
}

type ETSIConfig struct {
	CMMFConfigurationInformationLocator string                          `json:"cmmfConfigurationInformationLocator"`
	BitstreamVersion                    *int                            `json:"bitstreamVersion"`
	BitstreamID                         int                             `json:"bitstreamID"`
	ContentSourceType                   *string                         `json:"contentSourceType"`
	CodeType                            *int                            `json:"codeType"`
	Profile                             *string                         `json:"profile"`
	ProfileDescription                  *string                         `json:"profileDescription"`
	BlockSizeToBlockNumSymbolsMaps      []BlockSizeToBlockNumSymbolsMap `json:"blockSizeToBlockNumSymbolsMaps"`
	SymbolsPresentTarget                *int                            `json:"symbolsPresentTarget"`
	LiveStreamingBlockSymbolSize        *int                            `json:"liveStreamingBlockSymbolSize"`
	SymbolArrMap                        []SymbolArrangementMap          `json:"symbolArrangementMap"`
	XCDInfo                             XCDInformation                  `json:"xcdInformation"`
	BlockIntegrity                      *BlockIntegrity                 `json:"blockIntegrity"`
	PacketIntegrity                     *PacketIntegrity                `json:"packetIntegrity"`
}

type BlockSizeToBlockNumSymbolsMap struct {
	BlockSizeMin    *int `json:"blockSizeMin"`
	BlockSizeMax    *int `json:"blockSizeMax"`
	BlockNumSymbols *int `json:"blockNumSymbols"`
}

type SymbolArrangementMap struct {
	VariantID   *string `json:"variantId"`
	SymbolOrder []int   `json:"symbolOrder"`
}

type XCDInformation struct {
	BlockFieldSizeExponent  *int                   `json:"blockFieldSizeExponent"`
	CoefficientType         *string                `json:"coefficientType"`
	SignalSystematicSymbols *bool                  `json:"signalSystematicSymbols"`
	CoefficientInfo         CoefficientInformation `json:"coefficientInformation"`
	EncryptionParams        *EncryptionParameters  `json:"encryptionParameters"`
}

type BlockIntegrity struct {
	HashType      *string `json:"hashType"`
	HashAlgorithm *string `json:"hashAlgorithm"`
}

type PacketIntegrity struct {
	HashAlgorithm *string `json:"hashAlgorithm"`
	HashSize      *string `json:"hashSize"`
}

type CoefficientInformation struct {
	PRNGParams            []PrngParameters       `json:"prngParameters"`
	CoefficientVectorMaps []CoefficientVectorMap `json:"coefficientVectorMaps"`
}

type PrngParameters struct {
	VariantID             *string `json:"variantId"`
	BlockPRNGSeed         *int    `json:"blockPrngSeed"`
	PacketPRNGSeeds       []int   `json:"packetPrngSeeds"`
	PRNGDensityPercentage float64 `json:"prngDensityPercentage"`
}

type CoefficientVectorMap struct {
	VariantID          *string  `json:"variantId"`
	NumSymbols         *int     `json:"numSymbols"`
	FieldSizeExp       *int     `json:"fieldSizeExp"`
	CoefficientVectors []string `json:"coefficientVectors"`
}

type EncryptionParameters struct {
	EncryptionServiceEndpoint string `json:"encryptionServiceEndpoint"`
	BitstreamEncryptionKeyID  string `json:"bitstreamEncryptionKeyId"`
	BlockEncryptionMode       string `json:"blockEncryptionMode"`
	BlockKeySizeExp           *int   `json:"blockKeySizeExp"`
}

func NewValidatedETSIConfig(r io.Reader) (*ETSIConfig, error) {
	etsiConfig := &ETSIConfig{}

	err := json.NewDecoder(r).Decode(etsiConfig)
	if err != nil {
		return nil, err
	}

	// check for unsupported/incorrect fields
	if etsiConfig.BitstreamVersion == nil {
		return nil, ETSIConfigError{
			arg:     "bitstreamVersion",
			message: "Bitstream Version was required but not set",
		}
	}
	if *etsiConfig.BitstreamVersion != 0 {
		return nil, ETSIConfigError{
			arg:     "bitstreamVersion",
			message: fmt.Sprintf("value (%d) must be zero", *etsiConfig.BitstreamVersion),
		}
	}

	if etsiConfig.ContentSourceType == nil {
		defaultContentSourceType := "000b"
		etsiConfig.ContentSourceType = &defaultContentSourceType
	} else {
		contentSourceTypeSet := NewSet[string]([]string{"000b", "001b", "011b"})
		if !contentSourceTypeSet.Contains(*etsiConfig.ContentSourceType) {
			return nil, ETSIConfigError{
				arg: "contentSourceType",
				message: fmt.Sprintf("unsupported value (%s), must be one of [%s]",
					*etsiConfig.ContentSourceType, strings.Join([]string{"000b", "001b", "011b"}, ", ")),
			}
		}
	}

	if etsiConfig.CodeType == nil {
		return nil, ETSIConfigError{
			arg:     "codeType",
			message: "Code Type was required but not set",
		}
	}
	if *etsiConfig.CodeType != 0 {
		return nil, ETSIConfigError{
			arg:     "codeType",
			message: fmt.Sprintf("value (%d) must be zero", *etsiConfig.CodeType),
		}
	}

	if etsiConfig.Profile == nil {
		return nil, ETSIConfigError{
			arg:     "profile",
			message: "Profile was required but not set",
		}
	}
	if *etsiConfig.Profile != "org.etsi.cmmf.a" {
		return nil, ETSIConfigError{
			arg:     "profile",
			message: fmt.Sprintf("value (%s) must be 'org.etsi.cmmf.a'", *etsiConfig.Profile),
		}
	}

	for i, blkSzToBlkNumSymsMap := range etsiConfig.BlockSizeToBlockNumSymbolsMaps {
		if blkSzToBlkNumSymsMap.BlockSizeMin == nil ||
			blkSzToBlkNumSymsMap.BlockSizeMax == nil ||
			blkSzToBlkNumSymsMap.BlockNumSymbols == nil {
			return nil, ETSIConfigError{
				arg:     "blockSizeToBlockNumSymbolsMaps",
				message: fmt.Sprintf("missing one or more default fields in blockSizeToBlockNumSymbolsMap[%d]", i),
			}
		}
	}

	for i, symArrMap := range etsiConfig.SymbolArrMap {
		if symArrMap.VariantID == nil || *symArrMap.VariantID == "" ||
			len(symArrMap.SymbolOrder) == 0 || symArrMap.SymbolOrder == nil {
			return nil, ETSIConfigError{
				arg:     "symbolOrder",
				message: fmt.Sprintf("missing one or more default fields in symbolArrMap[%d]", i),
			}
		}
	}

	if etsiConfig.SymbolsPresentTarget != nil {
		if *etsiConfig.SymbolsPresentTarget < 0 || *etsiConfig.SymbolsPresentTarget > 100 {
			return nil, ETSIConfigError{
				arg: "symbolsPresentTarget",
				message: fmt.Sprintf("value (%d) must be between 0 and 100",
					*etsiConfig.SymbolsPresentTarget),
			}
		}
	}

	if etsiConfig.LiveStreamingBlockSymbolSize == nil {
		defaultLiveStreamingBlockSymbolSize := 10000
		etsiConfig.LiveStreamingBlockSymbolSize = &defaultLiveStreamingBlockSymbolSize
	}

	if etsiConfig.XCDInfo.BlockFieldSizeExponent == nil {
		defaultBlockFieldSizeExponent := 1
		etsiConfig.XCDInfo.BlockFieldSizeExponent = &defaultBlockFieldSizeExponent
	} else if *etsiConfig.XCDInfo.BlockFieldSizeExponent != 1 {
		return nil, ETSIConfigError{
			arg:     "blockFieldSizeExponent",
			message: fmt.Sprintf("value (%d): must be 1", *etsiConfig.XCDInfo.BlockFieldSizeExponent),
		}
	}

	if etsiConfig.XCDInfo.CoefficientType == nil {
		defaultCoefficientType := "vector"
		etsiConfig.XCDInfo.CoefficientType = &defaultCoefficientType
	} else {
		supportedCoeffTypes := []string{"block_prng", "packet_prng", "vector"}
		coeffTypeSet := NewSet[string](supportedCoeffTypes)

		if !coeffTypeSet.Contains(*etsiConfig.XCDInfo.CoefficientType) {
			return nil, ETSIConfigError{
				arg: "coefficientType",
				message: fmt.Sprintf("unsupported value (%s), must be one of [%s]",
					*etsiConfig.XCDInfo.CoefficientType,
					strings.Join(supportedCoeffTypes, ", ")),
			}
		}
	}

	if etsiConfig.XCDInfo.SignalSystematicSymbols == nil {
		defaultSignalSystematicSymbols := false
		etsiConfig.XCDInfo.SignalSystematicSymbols = &defaultSignalSystematicSymbols
	}

	for i, prngParam := range etsiConfig.XCDInfo.CoefficientInfo.PRNGParams {
		if prngParam.VariantID == nil ||
			(prngParam.BlockPRNGSeed == nil && (prngParam.PacketPRNGSeeds == nil || len(prngParam.PacketPRNGSeeds) == 0)) {
			return nil, ETSIConfigError{
				arg:     "prngParameters",
				message: fmt.Sprintf("missing one or more default fields in prngParameters[%d]", i),
			}
		}

		if prngParam.PRNGDensityPercentage < 0.0 ||
			prngParam.PRNGDensityPercentage > 1.0 {
			return nil, ETSIConfigError{
				arg: "coefficientInfo.PRNGParams.PRNGDensityPercentage",
				message: fmt.Sprintf("invalid value (%v) at coefficientInformation.prngParameters[%d], must be between 0.0 and 1.0",
					prngParam.PRNGDensityPercentage, i),
			}
		}
	}

	for i, coeffVecMap := range etsiConfig.XCDInfo.CoefficientInfo.CoefficientVectorMaps {
		if coeffVecMap.VariantID == nil || coeffVecMap.NumSymbols == nil ||
			coeffVecMap.FieldSizeExp == nil || coeffVecMap.CoefficientVectors == nil ||
			len(coeffVecMap.CoefficientVectors) == 0 {
			return nil, ETSIConfigError{
				arg:     "coefficientVectors",
				message: fmt.Sprintf("missing one or more default fields in coefficientVectorMaps[%d]", i),
			}
		}

		if *coeffVecMap.NumSymbols != len(coeffVecMap.CoefficientVectors) {
			return nil, ETSIConfigError{
				arg: "coefficientVectors",
				message: fmt.Sprintf("numSymbols (%d) must match the length of the coefficientVectors (%d)",
					*coeffVecMap.NumSymbols, len(coeffVecMap.CoefficientVectors)),
			}
		}
	}

	supportedFieldSizeExpValues := []int{1, 2, 4, 8, 16}
	fieldSizeExpSet := NewSet[int](supportedFieldSizeExpValues)
	for _, coefficientVectorMap := range etsiConfig.XCDInfo.CoefficientInfo.CoefficientVectorMaps {
		if !fieldSizeExpSet.Contains(*coefficientVectorMap.FieldSizeExp) {
			return nil, ETSIConfigError{
				arg: "fieldSizeExp",
				message: fmt.Sprintf("invalid value (%d), must be one of %v",
					*coefficientVectorMap.FieldSizeExp, supportedFieldSizeExpValues),
			}
		}
	}

	if etsiConfig.XCDInfo.EncryptionParams != nil {
		// todo: this should have a default value since it isn't a required field
		supportedBlockEncryptionModes := []string{"0000b", "0001b", "0010b", "0011b", "0100b", "0101b"}
		blockEncryptionModeSet := NewSet[string](supportedBlockEncryptionModes)
		if !blockEncryptionModeSet.Contains(etsiConfig.XCDInfo.EncryptionParams.BlockEncryptionMode) {
			return nil, ETSIConfigError{
				arg: "encryptionMode",
				message: fmt.Sprintf("unsupported encryption param value (%s), must be one of [%s]",
					etsiConfig.XCDInfo.EncryptionParams.BlockEncryptionMode,
					strings.Join(supportedBlockEncryptionModes, ", "),
				),
			}
		}

		if etsiConfig.XCDInfo.EncryptionParams.BlockKeySizeExp == nil {
			defaultBlockKeySizeExp := 8
			etsiConfig.XCDInfo.EncryptionParams.BlockKeySizeExp = &defaultBlockKeySizeExp
		}
	}

	if etsiConfig.BlockIntegrity != nil {
		if etsiConfig.BlockIntegrity.HashType == nil {
			defaultBlockIntegrityHashType := "00b"
			etsiConfig.BlockIntegrity.HashType = &defaultBlockIntegrityHashType
		} else {
			supportedBlockIntegrityHashTypes := []string{"00b", "01b"}
			blockIntegrityHashTypeSet := NewSet[string](supportedBlockIntegrityHashTypes)
			if !blockIntegrityHashTypeSet.Contains(*etsiConfig.BlockIntegrity.HashType) {
				return nil, ETSIConfigError{
					arg: "HashType",
					message: fmt.Sprintf("unsupported block integrity value (%s), must be one of [%s]",
						*etsiConfig.BlockIntegrity.HashType,
						strings.Join(supportedBlockIntegrityHashTypes, ", ")),
				}
			}
		}

		if etsiConfig.BlockIntegrity.HashAlgorithm == nil {
			defaultBlockIntegrityHashAlgorithm := "000b"
			etsiConfig.BlockIntegrity.HashAlgorithm = &defaultBlockIntegrityHashAlgorithm
		} else {
			supportedBlockIntegrityHashAlgos := []string{"000b", "001b", "010b"}
			blockIntegrityHashAlgoSet := NewSet[string](supportedBlockIntegrityHashAlgos)
			if !blockIntegrityHashAlgoSet.Contains(*etsiConfig.BlockIntegrity.HashAlgorithm) {
				return nil, ETSIConfigError{
					arg: "HashAlgorithm",
					message: fmt.Sprintf("unsupported block integrity value (%s), must be one of [%s]",
						*etsiConfig.BlockIntegrity.HashAlgorithm,
						strings.Join(supportedBlockIntegrityHashAlgos, ", ")),
				}
			}
		}
	}

	if etsiConfig.PacketIntegrity != nil {
		if etsiConfig.PacketIntegrity.HashAlgorithm == nil {
			defaultPacketIntegrityHashAlgorithm := "010b"
			etsiConfig.PacketIntegrity.HashAlgorithm = &defaultPacketIntegrityHashAlgorithm
		} else {
			supportedPacketIntegrityHashAlgos := []string{"000b", "001b", "010b"}
			packetIntegrityHashAlgoSet := NewSet[string](supportedPacketIntegrityHashAlgos)
			if !packetIntegrityHashAlgoSet.Contains(*etsiConfig.PacketIntegrity.HashAlgorithm) {
				return nil, ETSIConfigError{
					arg: "HashAlgorithm",
					message: fmt.Sprintf("unsupported packet integrity value (%s), must be one of [%s]",
						*etsiConfig.PacketIntegrity.HashAlgorithm,
						strings.Join(supportedPacketIntegrityHashAlgos, ", ")),
				}
			}
		}

		if etsiConfig.PacketIntegrity.HashSize == nil {
			defaultPacketIntegrityHashSize := "011b"
			etsiConfig.PacketIntegrity.HashSize = &defaultPacketIntegrityHashSize
		} else {
			supportedPacketIntegrityHashSizeSets := []string{"000b", "001b", "010b", "011b", "100b", "101b", "110b", "111b"}
			packetIntegrityHashSizeSet := NewSet[string](supportedPacketIntegrityHashSizeSets)
			if !packetIntegrityHashSizeSet.Contains(*etsiConfig.PacketIntegrity.HashSize) {
				return nil, ETSIConfigError{
					arg: "HashSize",
					message: fmt.Sprintf("unsupported packet integrity value (%s), must be one of [%s]",
						*etsiConfig.PacketIntegrity.HashSize,
						strings.Join(supportedPacketIntegrityHashSizeSets, ", ")),
				}
			}
		}
	}

	return etsiConfig, nil
}
