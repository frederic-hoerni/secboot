// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2026 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package efi

import (
	internal_efi "github.com/snapcore/secboot/internal/efi"
)

const (
	// allowSecurityLevelDowngradedParamKey is used to allow for the "Security Level is Downgraded to 0"
	// string in PCR7.
	allowSecurityLevelDowngradedParamKey loadParamsKey = "allow_security_level_downgraded"

	// includeSecurityLevelDowngradedParamKey is used to signal whether the "Security Level is Downgraded to 0"
	// string should be reflected in the produced PCR profile.
	// this is ignored if allowSecurityLevelDowngraded is false, as the presence of the event
	// will lead to an error in that case.
	includeSecurityLevelDowngradedParamKey = "include_security_level_downgraded"
)

type allowSecurityLevelDowngradedOption struct{}

func (o allowSecurityLevelDowngradedOption) ApplyOptionTo(visitor internal_efi.PCRProfileOptionVisitor) error {
	visitor.AddImageLoadParams(func(params ...loadParams) []loadParams {
		var out []loadParams
		for _, v := range []bool{false, true} {
			var newParams []loadParams
			for _, p := range params {
				newParams = append(newParams, p.Clone())
			}
			for _, p := range newParams {
				p[allowSecurityLevelDowngradedParamKey] = true
				p[includeSecurityLevelDowngradedParamKey] = v
			}
			out = append(out, newParams...)
		}
		return out
	})
	return nil
}

// WithAllowSecurityLevelDowngraded can be supplied to AddPCRProfile to allow for
// PCR7 including the "Security Level is Downgraded to 0" event. While this reduces security,
// it is required on some devices.
// If this string is present in the event log, this option results in a creation of a
// branched PCR profile that has two branches at the Firmware load stage one including
// the event with the string, the another not.
func WithAllowSecurityLevelDowngraded() PCRProfileOption {
	return allowSecurityLevelDowngradedOption{}
}
