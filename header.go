package webmention

import (
	"net/url"
)

type Link struct {
	URL    *url.URL
	Params map[string][]string
}

const (
	stateNormal int = iota
	stateParam
	stateParamValue
	stateURL
	stateEscape
	stateQuote
)

func GetHeaderLinks(headers []string) []*Link {
	links := make([]*Link, 0)
	for _, header := range headers {
		var link *Link
		state := stateNormal
		linkURL := ""
		paramName := ""
		paramVal := ""
		for _, c := range header {
			switch state {
			case stateNormal:
				if c == '<' {
					if link != nil {
						links = append(links, link)
					}
					link = &Link{Params: make(map[string][]string)}
					linkURL = ""
					state = stateURL
				} else if c == ';' {
					state = stateParam
				}
			case stateURL:
				if c != '>' {
					linkURL = linkURL + string(c)
				} else {
					link.URL, _ = url.Parse(linkURL)
					state = stateNormal
				}
			case stateParam:
				if c != '=' && c != ' ' {
					paramName = paramName + string(c)
				}
				if c == '=' {
					state = stateParamValue
				}
			case stateParamValue:
				if c == ' ' && paramVal == "" {
					continue
				}
				if c == ' ' {
					link.Params[paramName] = append(link.Params[paramName], paramVal)
					paramName = ""
					paramVal = ""
					state = stateNormal
					continue
				}
				if c == '"' && paramVal == "" {
					state = stateQuote
					continue
				}
				paramVal = paramVal + string(c)
			case stateQuote:
				if c == '\\' {
					state = stateEscape
					continue
				}
				if c == '"' {
					link.Params[paramName] = append(link.Params[paramName], paramVal)
					paramName = ""
					paramVal = ""
					state = stateNormal
					continue
				}
				paramVal = paramVal + string(c)
			case stateEscape:
				paramVal = paramVal + string(c)
				state = stateParamValue
			}
		}
		if link != nil {
			links = append(links, link)
		}
	}
	return links
}
