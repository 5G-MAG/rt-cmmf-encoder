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
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var sampleConfig = []byte(`
{
    "cmmfConfigurationInformationLocator": "https://www.example.com/cmmf_ci_manifest.mpd",
    "bitstreamVersion": 0,
    "contentSourceType": "001b",
    "codeType": 0,
    "profile": "org.etsi.cmmf.a",
	"symbolsPresentTarget": 10,
    "blockSizeToBlockNumSymbolsMaps": [
        {
            "blockSizeMin": 0,
            "blockSizeMax": 128,
            "blockNumSymbols": 1
        },
        {
            "blockSizeMin": 129,
            "blockSizeMax": 1048576,
            "blockNumSymbols": 4
        },
        {
            "blockSizeMin": 1048577,
            "blockSizeMax": 2097152,
            "blockNumSymbols": 8
        },
        {
            "blockSizeMin": 2097153,
            "blockSizeMax": 4194304,
            "blockNumSymbols": 16
        }
    ],
    "xcdInformation": {
		"blockFieldSizeExponent": 1,
        "signalSystematicSymbols": true,
        "coefficientInformation": {
            "coefficientVectorMaps": [
                {
                    "variantId": "cmmf-a",
                    "numSymbols": 4,
                    "fieldSizeExp": 1,
                    "coefficientVectors": [
                        "1000b",
                        "0010b",
                        "0101b",
                        "1111b"
                    ]
                },
                {
                    "variantId": "cmmf-b",
                    "numSymbols": 4,
                    "fieldSizeExp": 1,
                    "coefficientVectors": [
                        "0100b",
                        "0001b",
                        "1010b",
                        "1111b"
                    ]
                },
                {
                    "variantId": "cmmf-a",
                    "numSymbols": 8,
                    "fieldSizeExp": 1,
                    "coefficientVectors": [
                        "10000000b",
                        "00100000b",
                        "00001000b",
                        "00000010b",
                        "00000001b",
                        "00000100b",
                        "00010000b",
                        "11111111b"
                    ]
                },
                {
                    "variantId": "cmmf-b",
                    "numSymbols": 8,
                    "fieldSizeExp": 1,
                    "coefficientVectors": [
                        "01000000b",
                        "00010000b",
                        "00000100b",
                        "00000001b",
                        "00000010b",
                        "00001000b",
                        "00100000b",
                        "11111111b"
                    ]
                },
                {
                    "variantId": "cmmf-a",
                    "numSymbols": 16,
                    "fieldSizeExp": 1,
                    "coefficientVectors": [
                        "01000000b",
                        "00010000b",
                        "00000100b",
                        "00000001b",
                        "00000010b",
                        "00001000b",
                        "00100000b",
                        "11111111b",
                        "01000000b",
                        "00010000b",
                        "00000100b",
                        "00000001b",
                        "00000010b",
                        "00001000b",
                        "00100000b",
                        "11111111b"
                    ]
                },
                {
                    "variantId": "cmmf-b",
                    "numSymbols": 16,
                    "fieldSizeExp": 1,
                    "coefficientVectors": [
                        "01000000b",
                        "00010000b",
                        "00000100b",
                        "00000001b",
                        "00000010b",
                        "00001000b",
                        "00100000b",
                        "11111111b",
                        "01000000b",
                        "00010000b",
                        "00000100b",
                        "00000001b",
                        "00000010b",
                        "00001000b",
                        "00100000b",
                        "11111111b"
                    ]
                }
            ]
        }
    }
}`)

func Test_NewValidatedETSIConfig(t *testing.T) {
	tests := []struct {
		name           string
		encConfig      []byte
		modifiedFields bool
		updatedFields  map[string]any
		delFields      []string
		wantErr        bool
	}{
		{
			name:      "Happy case",
			encConfig: sampleConfig,
		},
		{
			name:           "Missing required fields (blockSizeMin) blockSizeToBlockNumSymbolsMaps",
			encConfig:      sampleConfig,
			modifiedFields: true,
			delFields:      []string{"BlockSizeToBlockNumSymbolsMaps.BlockSizeMin"},
			wantErr:        true,
		},
		{
			name:           "Missing required fields (blockSizeMax) blockSizeToBlockNumSymbolsMaps",
			encConfig:      sampleConfig,
			modifiedFields: true,
			delFields:      []string{"BlockSizeToBlockNumSymbolsMaps.BlockSizeMax"},
			wantErr:        true,
		},
		{
			name:           "Missing required fields (blockNumSymbols) blockSizeToBlockNumSymbolsMaps",
			encConfig:      sampleConfig,
			modifiedFields: true,
			delFields:      []string{"BlockSizeToBlockNumSymbolsMaps.BlockNumSymbols"},
			wantErr:        true,
		},
		{
			name:           "Invalid field (codeType)",
			encConfig:      sampleConfig,
			modifiedFields: true,
			updatedFields: map[string]any{
				"CodeType": 2,
			},
			wantErr: true,
		},
		{
			name:           "Invalid field (symbolsPresentTarget)",
			encConfig:      sampleConfig,
			modifiedFields: true,
			updatedFields: map[string]any{
				"SymbolsPresentTarget": 101,
			},
			wantErr: true,
		},
		{
			name:           "Invalid field (XCDInfo.BlockFieldSizeExponent)",
			encConfig:      sampleConfig,
			modifiedFields: true,
			updatedFields: map[string]any{
				"XCDInfo.BlockFieldSizeExponent": 101,
			},
			wantErr: true,
		},
		{
			name:           "Missing required fields (XCDInfo.CoefficientInfo.CoefficientVectorMaps.VariantID)",
			encConfig:      sampleConfig,
			modifiedFields: true,
			delFields:      []string{"XCDInfo.CoefficientInfo.CoefficientVectorMaps.VariantID"},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonBytes []byte
			etsiConfig := &ETSIConfig{}
			if tt.modifiedFields {
				_ = json.NewDecoder(bytes.NewReader(tt.encConfig)).Decode(etsiConfig)
				pointerVal := reflect.ValueOf(*etsiConfig)

				for _, k := range tt.delFields {
					sub := pointerVal
					keys := strings.Split(k, ".")
					for _, k := range keys[:len(keys)-1] {
						sub = sub.FieldByName(k)
						if sub.Kind() == reflect.Slice {
							sub = sub.Index(0)
						}
					}
					sub = sub.FieldByName(keys[len(keys)-1])
					sub.SetZero()
				}

				for k, v := range tt.updatedFields {
					sub := pointerVal
					keys := strings.Split(k, ".")
					for _, k := range keys[:len(keys)-1] {
						sub = sub.FieldByName(k)
						if sub.Kind() == reflect.Slice {
							sub = sub.Index(0)
						}
					}
					sub = sub.FieldByName(keys[len(keys)-1])
					if !sub.CanAddr() {
						sub = sub.Elem()
					}
					sub.Set(reflect.ValueOf(v))
				}

				jsonBytes, _ = json.Marshal(etsiConfig)
				fmt.Println(string(jsonBytes))
			} else {
				jsonBytes = tt.encConfig
			}

			encConfig, err := NewValidatedETSIConfig(bytes.NewReader(jsonBytes))

			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Equal(t, err, nil)
				assert.NotEqual(t, encConfig, nil)
			}
		})
	}
}
