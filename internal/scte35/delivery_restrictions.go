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

const (
	DeviceRestrictionsGroup0 uint32 = iota
	DeviceRestrictionsGroup1
	DeviceRestrictionsGroup2
	DeviceRestrictionsNone
)

// DeliveryRestrictions contains the specific delivery restriction flags as
// defined within the SegmentationDescriptorType XML schema definition.
type DeliveryRestrictions struct {
	ArchiveAllowedFlag     bool
	WebDeliveryAllowedFlag bool
	NoRegionalBlackoutFlag bool
	DeviceRestrictions     uint32
}

// deviceRestrictionsName returns the human-readable name for the
// device_restrictions.
func (dr *DeliveryRestrictions) deviceRestrictionsName() string {
	switch dr.DeviceRestrictions {
	case DeviceRestrictionsGroup0:
		return "Restrict Group 0"
	case DeviceRestrictionsGroup1:
		return "Restrict Group 1"
	case DeviceRestrictionsGroup2:
		return "Restrict Group 2"
	case DeviceRestrictionsNone:
		return "None"
	default:
		return "Unknown"
	}
}
