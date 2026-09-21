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
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func bitstreamHeader_getETSIConfigWithProfile(profile, profileDescription string) *ETSIConfig {
	contentSourceType := "001b"
	return &ETSIConfig{
		Profile:            &profile,
		ProfileDescription: &profileDescription,
		ContentSourceType:  &contentSourceType,
	}
}

func TestBitstreamHeader_Serialize(t *testing.T) {
	type fields struct {
		contentSourceSize uint64
		encoderConfig     *ETSIConfig
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "Arbitrary profile string",
			fields: fields{
				contentSourceSize: 0,
				encoderConfig:     bitstreamHeader_getETSIConfigWithProfile("DOLBY", ""),
			},
			want: []byte{
				0x20, 0x15, //subatom header
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // source size
				0x20,                               // content_source_type, content_source_split & code_type (first 3 bits)
				0x00,                               // code_type (last bit), rfc_5052, block_count_minus1 (first 6 bits)
				0x15, 0x44, 0x4f, 0x4c, 0x42, 0x59, // Profile info (offset by 3 bits)
				0x00, 0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "'org.etsi.cmmf.a' Profile",
			fields: fields{
				contentSourceSize: 0,
				encoderConfig:     bitstreamHeader_getETSIConfigWithProfile("org.etsi.cmmf.a", ""),
			},
			want: []byte{
				0x20, 0x1f, //subatom header
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // source size
				0x20,                                           // content_source_type, content_source_split & code_type (first 3 bits)
				0x00,                                           // code_type (last bit), rfc_5052, block_count_minus1 (first 6 bits)
				0x1f, 0x6f, 0x72, 0x67, 0x2e, 0x65, 0x74, 0x73, // Profile info
				0x69, 0x2e, 0x63, 0x6d, 0x6d, 0x66, 0x2e, 0x61,
				0x00, 0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "Size zero",
			fields: fields{
				contentSourceSize: 0,
				encoderConfig:     bitstreamHeader_getETSIConfigWithProfile("org.etsi.cmmf.a", ""),
			},
			want: []byte{
				0x20, 0x1f, //subatom header
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // source size
				0x20, // content_source_type, content_source_split & code_type (first 3 bits)
				0x00, // code_type (last bit), rfc_5052, block_count_minus1 (first 6 bits)
				0x1f, 0x6f, 0x72, 0x67, 0x2e, 0x65, 0x74, 0x73, 0x69, 0x2e, 0x63, 0x6d, 0x6d, 0x66, 0x2e, 0x61,
				0x0, 0x0, 0x0, 0x0, 0x0,
			},
		},
		{
			name: "Max size",
			fields: fields{
				contentSourceSize: math.MaxUint64,
				encoderConfig:     bitstreamHeader_getETSIConfigWithProfile("org.etsi.cmmf.a", ""),
			},
			want: []byte{
				0x20, 0x1f, //subatom header
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // source size
				0x20, // content_source_type, content_source_split & code_type (first 3 bits)
				0x00, // code_type (last bit), rfc_5052, block_count_minus1 (first 6 bits)
				0x1f, 0x6f, 0x72, 0x67, 0x2e, 0x65, 0x74, 0x73, 0x69, 0x2e, 0x63, 0x6d, 0x6d, 0x66, 0x2e, 0x61,
				0x0, 0x0, 0x0, 0x0, 0x0,
			},
		},
		{
			name: "Random size",
			fields: fields{
				contentSourceSize: 2831278340,
				encoderConfig:     bitstreamHeader_getETSIConfigWithProfile("org.etsi.cmmf.a", ""),
			},
			want: []byte{
				0x20, 0x1f, //subatom header
				0x00, 0x00, 0x00, 0x00, 0xa8, 0xc1, 0xe1, 0x04, // source size
				0x20, // content_source_type, content_source_split & code_type (first 3 bits)
				0x00, // code_type (last bit), rfc_5052, block_count_minus1 (first 6 bits)
				0x1f, 0x6f, 0x72, 0x67, 0x2e, 0x65, 0x74, 0x73, 0x69, 0x2e, 0x63, 0x6d, 0x6d, 0x66, 0x2e, 0x61,
				0x0, 0x0, 0x0, 0x0, 0x0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bitstreamHeaderSubatom := NewBitstreamHeaderFromEncoderConfig(tt.fields.encoderConfig, tt.fields.contentSourceSize)
			serialized := bitstreamHeaderSubatom.Serialize()
			assert.Equal(t, tt.want, serialized)
		})
	}

}
