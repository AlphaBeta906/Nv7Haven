package eod

import (
	"encoding/json"
	"time"
)

var starterElements = []element{
	{
		Name:       "Air",
		Category:   "",
		Comment:    "The invisible gaseous substance surrounding the earth, a mixture mainly of oxygen and nitrogen.",
		Creator:    "",
		Complexity: 0,
		Parents:    []string{},
	},
	{
		Name:       "Earth",
		Category:   "",
		Comment:    "The substance of the land surface; soil.",
		Creator:    "",
		Complexity: 0,
		Parents:    []string{},
	},
	{
		Name:       "Fire",
		Category:   "",
		Comment:    "Combustion or burning, in which substances combine chemically with oxygen from the air and typically give out bright light, heat, and smoke.",
		Creator:    "",
		Complexity: 0,
		Parents:    []string{},
	},
	{
		Name:       "Water",
		Category:   "",
		Comment:    "A colorless, transparent, odorless liquid that forms the seas, lakes, rivers, and rain and is the basis of the fluids of living organisms.",
		Creator:    "",
		Complexity: 0,
		Parents:    []string{},
	},
}

func (b *EoD) checkServer(m msg, rsp rsp) bool {
	lock.RLock()
	dat, exists := b.dat[m.GuildID]
	lock.RUnlock()
	if !exists {
		rsp.ErrorMessage("No voting or news channel has been set!")
		return false
	}
	if dat.votingChannel == "" {
		rsp.ErrorMessage("No voting channel has been set!")
		return false
	}
	if dat.newsChannel == "" {
		rsp.ErrorMessage("No news channel has been set!")
		return false
	}
	if dat.elemCache == nil {
		dat.elemCache = make(map[string]element)
	}
	if len(dat.elemCache) < 4 {
		for _, elem := range starterElements {
			elem.Guild = m.GuildID
			elem.CreatedOn = time.Now()
			dat.elemCache[elem.Name] = elem
			_, err := b.db.Exec("INSERT INTO eod_elements VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ? )", elem.Name, elem.Category, elem.Guild, elem.Comment, elem.Creator, int(elem.CreatedOn.Unix()), "", "", elem.Complexity)
			rsp.Error(err)
		}
		lock.Lock()
		b.dat[m.GuildID] = dat
		lock.Unlock()
	}

	if dat.invCache == nil {
		dat.invCache = make(map[string]map[string]empty)
	}
	_, exists = dat.invCache[m.Author.ID]
	if !exists {
		dat.invCache[m.Author.ID] = make(map[string]empty)
		for _, val := range starterElements {
			dat.invCache[m.Author.ID][val.Name] = empty{}
		}
		data, err := json.Marshal(dat.invCache[m.Author.ID])
		if rsp.Error(err) {
			return false
		}
		_, err = b.db.Exec("INSERT INTO eod_inv VALUES ( ?, ?, ?, ? )", m.GuildID, m.Author.ID, string(data), len(dat.invCache[m.Author.ID]))
		rsp.Error(err)
		lock.Lock()
		b.dat[m.GuildID] = dat
		lock.Unlock()
	}

	_, exists = dat.playChannels[m.ChannelID]
	return exists
}
