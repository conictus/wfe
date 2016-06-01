package wfe

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/url"
	"strings"
)

var (
	MgoGraphCollection = "graph"
)

type mgoGraphModel struct {
	ID       string   `bson:"_id"`
	ParentID string   `bson:"parent_id"`
	Function string   `bson:"function"`
	Args     []string `bson:"args"`
	State    string   `bson:"state"`
	Error    string   `bson:"error"`
	Result   string   `bson:"result"`
}

type mgoGrapher struct {
	session *mgo.Session
	db      string
}

type mgoGraph struct {
	id      string
	session *mgo.Session
	db      string
}

func init() {
	indexies := []mgo.Index{
		{
			Key: []string{"parent_id"},
		},
		{
			Key: []string{"state"},
		},
	}

	RegisterGraphBackend("mongodb", func(u *url.URL) (GraphBackend, error) {
		session, err := mgo.Dial(u.String())
		if err != nil {
			return nil, err
		}

		db := strings.Trim(u.Path, "/")
		//ensure index.
		c := session.DB(db).C(MgoGraphCollection)
		for _, index := range indexies {
			if err := c.EnsureIndex(index); err != nil {
				return nil, err
			}
		}

		return &mgoGrapher{
			session: session,
			db:      db,
		}, nil
	})
}

func (g *mgoGrapher) stringify(l []interface{}) []string {
	var r []string
	for _, i := range l {
		r = append(r, fmt.Sprintf("%v", i))
	}

	return r
}

func (g *mgoGrapher) Graph(id string, request Request) (Graph, error) {
	s := g.session.Copy()
	defer s.Close()

	_, err := s.DB(g.db).C(MgoGraphCollection).UpsertId(id, &mgoGraphModel{
		ID:       id,
		ParentID: request.ParentID(),
		Function: request.Fn(),
		State:    "running",
		Args:     g.stringify(request.Args()),
	})

	if err != nil {
		return nil, err
	}

	return &mgoGraph{
		id:      id,
		session: g.session,
		db:      g.db,
	}, nil
}

func (g *mgoGraph) Commit(response *Response) error {
	s := g.session.Copy()
	defer s.Close()

	return s.DB(g.db).C(MgoGraphCollection).UpdateId(
		g.id,
		bson.M{
			"$set": bson.M{
				"state":  response.State,
				"error":  response.Error,
				"result": fmt.Sprintf("%v", response.Result),
			},
		})
}
