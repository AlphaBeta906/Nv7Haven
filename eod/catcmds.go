package eod

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Nv7-Github/Nv7Haven/eod/types"
	"github.com/Nv7-Github/Nv7Haven/eod/util"
	"github.com/bwmarrin/discordgo"
)

const x = "❌"
const check = "✅"

func (b *EoD) catCmd(category string, sortKind string, hasUser bool, user string, m types.Msg, rsp types.Rsp) {
	lock.RLock()
	dat, exists := b.dat[m.GuildID]
	lock.RUnlock()
	if !exists {
		return
	}

	category = strings.TrimSpace(category)

	if isFoolsMode && !isFool(category) {
		rsp.ErrorMessage(makeFoolResp(category))
		return
	}

	id := m.Author.ID
	if hasUser {
		id = user
	}
	inv, res := dat.GetInv(id, !hasUser)
	if !res.Exists {
		rsp.ErrorMessage(res.Message)
		return
	}

	cat, res := dat.GetCategory(category)
	if !res.Exists {
		rsp.ErrorMessage(res.Message)
		return
	}
	category = cat.Name

	out := make([]struct {
		found int
		text  string
		name  string
	}, len(cat.Elements))

	found := 0
	i := 0
	fnd := 0
	var text string

	for name := range cat.Elements {
		_, exists := inv[strings.ToLower(name)]
		if exists {
			text = name + " " + check
			found++
			fnd = 1
		} else {
			text = name + " " + x
			fnd = 0
		}

		out[i] = struct {
			found int
			text  string
			name  string
		}{
			found: fnd,
			text:  text,
			name:  name,
		}

		i++
	}

	switch sortKind {
	case "catfound":
		sort.Slice(out, func(i, j int) bool {
			return out[i].found > out[j].found
		})

	case "catnotfound":
		sort.Slice(out, func(i, j int) bool {
			return out[i].found < out[j].found
		})

	case "catelemcount":
		rsp.ErrorMessage("Invalid sort!")
		return

	default:
		sorter := sorts[sortKind]
		dat.Lock.RLock()
		sort.Slice(out, func(i, j int) bool {
			return sorter(out[i].name, out[j].name, dat)
		})
		dat.Lock.RUnlock()
	}

	o := make([]string, len(out))
	for i, val := range out {
		o[i] = val.text
	}

	b.newPageSwitcher(types.PageSwitcher{
		Kind:       types.PageSwitchInv,
		Thumbnail:  cat.Image,
		Title:      fmt.Sprintf("%s (%d, %s%%)", category, len(out), util.FormatFloat(float32(found)/float32(len(out))*100, 2)),
		PageGetter: b.invPageGetter,
		Items:      o,
	}, m, rsp)
}

type catData struct {
	text  string
	name  string
	found float32
	count int
}

func (b *EoD) allCatCmd(sortBy string, hasUser bool, user string, m types.Msg, rsp types.Rsp) {
	lock.RLock()
	dat, exists := b.dat[m.GuildID]
	lock.RUnlock()
	if !exists {
		return
	}

	id := m.Author.ID
	if hasUser {
		id = user
	}
	inv, res := dat.GetInv(id, !hasUser)
	if !res.Exists {
		rsp.ErrorMessage(res.Message)
		return
	}

	dat.Lock.RLock()
	out := make([]catData, len(dat.Categories))

	i := 0
	for _, cat := range dat.Categories {
		count := 0
		for elem := range cat.Elements {
			_, exists := inv[strings.ToLower(elem)]
			if exists {
				count++
			}
		}

		perc := float32(count) / float32(len(cat.Elements))
		text := "(" + util.FormatFloat(perc*100, 2) + "%)"
		if count == len(cat.Elements) {
			text = check
		}
		out[i] = catData{
			text:  fmt.Sprintf("%s %s", cat.Name, text),
			name:  cat.Name,
			found: perc,
			count: len(cat.Elements),
		}
		i++
	}
	dat.Lock.RUnlock()

	switch sortBy {
	case "catfound":
		sort.Slice(out, func(i, j int) bool {
			return out[i].found > out[j].found
		})

	case "catnotfound":
		sort.Slice(out, func(i, j int) bool {
			return out[i].found < out[j].found
		})

	case "catelemcount":
		sort.Slice(out, func(i, j int) bool {
			return out[i].count > out[j].count
		})

	default:
		sort.Slice(out, func(i, j int) bool {
			return compareStrings(out[i].name, out[j].name)
		})
	}

	names := make([]string, len(out))
	for i, dat := range out {
		names[i] = dat.text
	}

	b.newPageSwitcher(types.PageSwitcher{
		Kind:       types.PageSwitchInv,
		Title:      fmt.Sprintf("All Categories (%d)", len(out)),
		PageGetter: b.invPageGetter,
		Items:      names,
	}, m, rsp)
}

func (b *EoD) downloadCatCmd(catName string, sort string, m types.Msg, rsp types.Rsp) {
	lock.RLock()
	dat, exists := b.dat[m.GuildID]
	lock.RUnlock()
	if !exists {
		return
	}

	cat, res := dat.GetCategory(catName)
	if !res.Exists {
		rsp.ErrorMessage(res.Message)
		return
	}

	dat.Lock.RLock()
	elems := make([]string, len(cat.Elements))
	i := 0

	for elem := range cat.Elements {
		elems[i] = elem
		i++
	}
	dat.Lock.RUnlock()

	sortElemList(elems, sort, dat)

	out := &strings.Builder{}
	for _, elem := range elems {
		out.WriteString(elem + "\n")
	}
	buf := strings.NewReader(out.String())

	channel, err := b.dg.UserChannelCreate(m.Author.ID)
	if rsp.Error(err) {
		return
	}

	b.dg.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Content: fmt.Sprintf("Category **%s**:", cat.Name),
		Files: []*discordgo.File{
			{
				Name:        "cat.txt",
				ContentType: "text/plain",
				Reader:      buf,
			},
		},
	})
	rsp.Message("Sent category in DMs!")
}
