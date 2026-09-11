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
	"testing"

	"github.com/stretchr/testify/assert"
)

func sync_getETSIConfigWithVersion(version int) *ETSIConfig {
	bitstreamVersion := version
	return &ETSIConfig{
		BitstreamVersion: &bitstreamVersion,
	}
}

func TestSync_SyncFromConfig(t *testing.T) {
	type fields struct {
		encoderConfig *ETSIConfig
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "Happy path",
			fields: fields{
				encoderConfig: sync_getETSIConfigWithVersion(0),
			},
			want: []byte{0x89, 0x78, 0x0D, 0x43, 0x00, 0x44, 0xAC, 0x31, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncSubatom := NewSyncFromEncoderConfig(tt.fields.encoderConfig)
			serialized := syncSubatom.Serialize()
			assert.Equal(t, tt.want, serialized)
		})
	}

}
