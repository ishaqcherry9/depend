package jy2struct

import (
	"bytes"
	"errors"
	"os"
	"strings"
)

type Args struct {
	Format    string
	Data      string
	InputFile string
	Name      string
	SubStruct bool
	Tags      string

	tags          []string
	convertFloats bool
	parser        Parser
}

func (j *Args) checkValid() error {
	switch j.Format {
	case "json":
		j.parser = ParseJSON
		j.convertFloats = true
	case "yaml":
		j.parser = ParseYaml
	default:
		return errors.New("format must be json or yaml")
	}

	j.tags = []string{j.Format}
	tags := strings.Split(j.Tags, ",")
	for _, tag := range tags {
		if tag == j.Format || tag == "" {
			continue
		}
		j.tags = append(j.tags, tag)
	}

	if j.Name == "" {
		j.Name = "GenerateName"
	}

	return nil
}

func Convert(args *Args) (string, error) {
	err := args.checkValid()
	if err != nil {
		return "", err
	}

	var data []byte
	if args.Data != "" {
		data = []byte(args.Data)
	} else {
		data, err = os.ReadFile(args.InputFile)
		if err != nil {
			return "", err
		}
	}

	input := bytes.NewReader(data)

	output, err := jyParse(input, args.parser, args.Name, "main", args.tags, args.SubStruct, args.convertFloats)
	if err != nil {
		return "", err
	}

	return string(output), nil
}
