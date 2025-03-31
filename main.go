/*
 * Copyright (C) 2025 Gloria Ciavarrini
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"fmt"

	"korifi-client/korifi"
)

func main() {
	client, err := korifi.GetKorifiHttpClient()
	if err != nil {
		fmt.Printf("Error creating HTTP client: %v\n", err)
		return
	}

	info, err := korifi.GetInfo(client)
	if err != nil {
		fmt.Printf("Error getting info: %v\n", err)
		return
	}

	fmt.Printf("Korifi Info: %+v\n", info)
}
