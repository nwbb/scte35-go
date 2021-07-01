// Copyright 2021 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or   implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package scte35

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/bamiaux/iobit"
)

const (
	// TimeDescriptorTag is the splice_descriptor_tag for a time_descriptor.
	TimeDescriptorTag = 0x03
)

// TimeDescriptor is an implementation of a splice_descriptor. It provides an
// optional extension to the splice_insert(), splice_null() and time_signal()
// commands that allows a programmer’s wall clock time to be sent to a client.
// For the highest accuracy, this descriptor should be used with the
// time_signal() or splice_insert( ) command. This command may be inserted using
// SCTE 104 or by out of band provisioning on the device inserting this message.
type TimeDescriptor struct {
	XMLName    xml.Name `xml:"TimeDescriptor" json:"-"`
	JSONType   uint32   `xml:"-" json:"type"`
	TAISeconds uint64   `xml:"taiSeconds,attr" json:"taiSeconds"`
	TAINS      uint32   `xml:"taiNs,attr" json:"taiNs"`
	UTCOffset  uint32   `xml:"utcOffset,attr" json:"utcOffset"`
}

// Tag returns the splice_descriptor_tag.
func (sd *TimeDescriptor) Tag() uint32 {
	// ensure JSONType is set
	sd.JSONType = TimeDescriptorTag
	return TimeDescriptorTag
}

// String returns a table description of this splice_descriptor.
func (sd *TimeDescriptor) String() string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("time_descriptor() {\n"))
	buf.WriteString(fmt.Sprintf("    splice_descriptor_tag: %#02x\n", TimeDescriptorTag))
	buf.WriteString(fmt.Sprintf("    descriptor_length: %d bytes\n", sd.length()))
	buf.WriteString(fmt.Sprintf("    identifier: %s\n", CUEIASCII))
	buf.WriteString(fmt.Sprintf("    TAI_seconds: %d", sd.TAISeconds))
	buf.WriteString(fmt.Sprintf("    TAI_ns: %d\n", sd.TAINS))
	buf.WriteString(fmt.Sprintf("    UTC_offset: %d\n", sd.UTCOffset))
	buf.WriteString(fmt.Sprintf("}\n"))

	return buf.String()
}

// decode updates this splice_descriptor from binary.
func (sd *TimeDescriptor) decode(b []byte) error {
	r := iobit.NewReader(b)
	r.Skip(8)  // splice_descriptor_tag
	r.Skip(8)  // descriptor_length
	r.Skip(32) // identifier
	sd.TAISeconds = r.Uint64(48)
	sd.TAINS = r.Uint32(32)
	sd.UTCOffset = r.Uint32(16)

	return readerError(r)
}

// encode this splice_descriptor to binary.
func (sd *TimeDescriptor) encode() ([]byte, error) {
	length := sd.length()

	// add 2 bytes to contain splice_descriptor_tag & descriptor_length
	buf := make([]byte, length+2)
	iow := iobit.NewWriter(buf)
	iow.PutUint32(8, TimeDescriptorTag)
	iow.PutUint32(8, uint32(length))
	iow.PutUint32(32, CUEIdentifier)
	iow.PutUint64(48, sd.TAISeconds)
	iow.PutUint32(32, sd.TAINS)
	iow.PutUint32(16, sd.UTCOffset)

	err := iow.Flush()
	return buf, err
}

// descriptorLength returns descriptor_length.
func (sd *TimeDescriptor) length() int {
	length := 32 // identifier
	length += 48 // TAI_seconds
	length += 32 // TAI_ns
	length += 16 // UTC_offset
	return length / 8
}
