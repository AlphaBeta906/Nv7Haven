package elemental

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v2"

	firebase "firebase.google.com/go"
	fire "github.com/Nv7-Github/firebase"
	authentication "github.com/Nv7-Github/firebase/auth"
	database "github.com/Nv7-Github/firebase/db"
	"google.golang.org/api/option"
)

var db *database.Db
var store *firestore.Client
var auth *authentication.Auth

// Element has the data for a created element
type Element struct {
	Color     string   `json:"color"`
	Comment   string   `json:"comment"`
	CreatedOn int      `json:"createdOn"`
	Creator   string   `json:"creator"`
	Name      string   `json:"name"`
	Parents   []string `json:"parents"`
	Pioneer   string   `json:"pioneer"`
}

// Color has the data for a suggestion's color
type Color struct {
	Base       string  `json:"base"`
	Lightness  float32 `json:"lightness"`
	Saturation float32 `json:"saturation"`
}

// Suggestion has the data for a suggestion
type Suggestion struct {
	Creator string   `json:"creator"`
	Name    string   `json:"name"`
	Votes   int      `json:"votes"`
	Color   Color    `json:"color"`
	Voted   []string `json:"voted"`
}

// ComboMap has the data that maps combos
type ComboMap map[string]map[string]string

// SuggMap has the data that maps suggestion combos
type SuggMap map[string]map[string][]string

// Recent has the data of a recent element
type Recent struct {
	Parents [2]string `json:"recipe"`
	Result  string    `json:"result"`
}

// InitElemental initializes all of Elemental's handlers on the app.
func InitElemental(app *fiber.App) error {
	opt := option.WithCredentialsJSON([]byte(serviceAccount))
	config := &firebase.Config{
		DatabaseURL:   "https://elementalserver-8c6d0.firebaseio.com",
		ProjectID:     "elementalserver-8c6d0",
		StorageBucket: "elementalserver-8c6d0.appspot.com",
	}
	fireapp, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		return err
	}

	firebaseapp, err := fire.CreateAppWithServiceAccount("https://elementalserver-8c6d0.firebaseio.com", "AIzaSyCsqvV3clnwDTTgPHDVO2Yatv5JImSUJvU", []byte(serviceAccount))
	if err != nil {
		return err
	}
	auth = authentication.CreateAuth(firebaseapp)

	db = database.CreateDatabase(firebaseapp)

	store, err = fireapp.Firestore(context.Background())
	if err != nil {
		return err
	}

	app.Get("/get_combo/:elem1/:elem2", getCombo)
	app.Get("/get_elem/:elem", getElem)
	app.Get("/get_found/:uid", getFound)
	app.Get("/new_found/:uid/:elem", newFound)
	app.Get("/clear", func(c *fiber.Ctx) error {
		cache = make(map[string]Element, 0)
		elemMap = make(map[string]map[string]string, 0)
		return nil
	})
	return nil
}

// CloseElemental has the cleanup functions
func CloseElemental() {
	store.Close()
}
