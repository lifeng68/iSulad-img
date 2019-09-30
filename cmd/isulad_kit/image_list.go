// Copyright (c) Huawei Technologies Co., Ltd. 2019-2019. All rights reserved.
// iSulad-kit licensed under the Mulan PSL v1.
// You can use this software according to the terms and conditions of the Mulan PSL v1.
// You may obtain a copy of Mulan PSL v1 at:
//     http://license.coscl.org.cn/MulanPSL
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v1 for more details.
// Description: iSulad image kit
// Author: lifeng
// Create: 2019-05-06

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/containers/image/types"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type listImagesResponse struct {
	// List of images.
	Images []*Image `json:"images,omitempty"`
}

// getUserFromImage gets uid or user name
func getUserFromImage(user string) (*int64, string) {
	if user == "" {
		return nil, ""
	}

	user = strings.Split(user, ":")[0]
	uid, err := strconv.ParseInt(user, 10, 64)
	if err != nil {
		return nil, user
	}

	return &uid, ""
}

func imagesHandler(c *cli.Context) error {

	filter := ""
	if c.IsSet("filter") {
		filter = c.String("filter")
	}

	store, err := getStorageStore(true, c)
	if err != nil {
		return err
	}

	ctx, cancel := commandTimeoutContextFromGlobalOptions(c)
	defer cancel()

	imageService, err := getImageService(ctx, c, store)
	if err != nil {
		return err
	}

	if c.Bool("check") {
		err = imageService.CheckImages(&types.SystemContext{})
		if err != nil {
			return err
		}
	}

	results, err := imageService.ListImages(&types.SystemContext{}, filter)
	if err != nil {
		return err
	}

	resp := &listImagesResponse{}
	for _, result := range results {
		imageConfig, err := getImageConf(store, result.ID)
		if err != nil {
			return err
		}
		healthcheck, err := getHealthcheck(store, result.ID)
		if err != nil {
			return err
		}
		resImg := &Image{
			ID:          result.ID,
			RepoTags:    result.RepoTags,
			RepoDigests: result.RepoDigests,
			Created:     result.Created,
			Loaded:      result.Loaded,
			ImageSpec:   imageConfig,
			Healthcheck: healthcheck,
		}
		uid, username := getUserFromImage(result.User)
		if uid != nil {
			resImg.UID = &Int64Value{Value: *uid}
		}
		resImg.Username = username
		if result.Size != nil {
			resImg.Size = *result.Size
		}
		resp.Images = append(resp.Images, resImg)
	}

	logrus.Debugf("listImagesResponse: %+v", resp)

	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", data)
	return err
}

var imagesCmd = cli.Command{
	Name:  "images",
	Usage: "isulad_kit images [FILTER]",
	Description: fmt.Sprintf(`

	List images.

	`),
	ArgsUsage: "images",
	Action:    imagesHandler,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "filter",
			Usage: "Filter output based on conditions provided",
		},
		cli.BoolFlag{
			Name:  "check",
			Usage: "enable check image integrity",
		},
	},
}
