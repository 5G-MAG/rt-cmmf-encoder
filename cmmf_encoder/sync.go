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

type Sync struct {
	SubatomID     SubatomID
	encoderConfig *ETSIConfig
}

func NewSyncFromEncoderConfig(encoderConfig *ETSIConfig) *Sync {
	return &Sync{
		SubatomID:     SUBATOM_ID_SYNC,
		encoderConfig: encoderConfig,
	}
}

func (sync *Sync) Serialize() []byte {
	serializedSyncSubatom := make([]byte, len(CMMFSyncWord))

	offset := copy(serializedSyncSubatom, CMMFSyncWord[:])

	serializedSyncSubatom = append(serializedSyncSubatom, 0)
	serializedSyncSubatom[offset] = serializedSyncSubatom[offset] | (uint8(*sync.encoderConfig.BitstreamVersion) << 4)

	return serializedSyncSubatom
}
