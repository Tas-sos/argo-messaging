package stores

import (
	"errors"
	"time"
)

// MockStore holds configuration
type MockStore struct {
	Server      string
	Database    string
	SubList     []QSub
	TopicList   []QTopic
	ProjectList []QProject
	UserList    []QUser
	RoleList    []QRole
	Session     bool
	TopicsACL   map[string]QAcl
	SubsACL     map[string]QAcl
}

// QueryACL Topic/Subscription ACL
func (mk *MockStore) QueryACL(projectUUID string, resource string, name string) (QAcl, error) {
	if resource == "topic" {
		if _, exists := mk.TopicsACL[name]; exists {
			return mk.TopicsACL[name], nil
		}
	} else if resource == "subscription" {
		if _, exists := mk.SubsACL[name]; exists {
			return mk.SubsACL[name], nil
		}
	}

	return QAcl{}, errors.New("not found")
}

// NewMockStore creates new mock store
func NewMockStore(server string, database string) *MockStore {
	mk := MockStore{}
	mk.Server = server
	mk.Database = database
	mk.Session = true
	mk.Initialize()
	return &mk
}

// Close is used to close session
func (mk *MockStore) Close() {
	mk.Session = false
}

// InsertUser inserts a new user to the store
func (mk *MockStore) InsertUser(uuid string, projects []QProjectRoles, name string, token string, email string, serviceRoles []string) error {
	user := QUser{UUID: uuid, Name: name, Email: email, Projects: projects, Token: token, ServiceRoles: serviceRoles}
	mk.UserList = append(mk.UserList, user)
	return nil
}

// UpdateUser updates user information
func (mk *MockStore) UpdateUser(uuid string, projects []QProjectRoles, name string, token string, email string, serviceAdmin bool) error {

	for i, item := range mk.UserList {
		if item.UUID == uuid {
			if projects != nil {
				if len(projects) > 0 {
					mk.UserList[i].Projects = projects
				}
			}
			if name != "" {
				mk.UserList[i].Name = name
			}
			if email != "" {
				mk.UserList[i].Email = email
			}

			return nil
		}
	}

	return errors.New("not found")

}

// HasUsers accepts a user array of usernames and returns the not found
func (mk *MockStore) HasUsers(projectUUID string, users []string) (bool, []string) {

	var notFound []string

	// for each given username
	for _, username := range users {
		found := false
		// loop through all found users
		for _, user := range mk.UserList {
			if username == user.Name {
				found = true
			}
		}
		// if not found add it to the notFound
		if !found {
			notFound = append(notFound, username)
		}

	}

	return len(notFound) == 0, notFound
}

// ModACL changes the acl in a function
func (mk *MockStore) ModACL(projectUUID string, resource string, name string, acl []string) error {
	newAcl := QAcl{ACL: acl}
	if resource == "topics" {
		if _, exists := mk.TopicsACL[name]; exists {
			mk.TopicsACL[name] = newAcl
			return nil
		}
	} else if resource == "subscriptions" {
		if _, exists := mk.SubsACL[name]; exists {
			mk.SubsACL[name] = newAcl
			return nil
		}
	}

	return errors.New("not found")
}

// UpdateProject updates project information
func (mk *MockStore) UpdateProject(projectUUID string, name string, description string, modifiedOn time.Time) error {

	for i, item := range mk.ProjectList {
		if item.UUID == projectUUID {
			if description != "" {
				mk.ProjectList[i].Description = description
			}
			if name != "" {
				mk.ProjectList[i].Name = name
			}

			mk.ProjectList[i].ModifiedOn = modifiedOn
			return nil
		}
	}

	return errors.New("not found")

}

// UpdateSubOffset updates the offset of the current subscription
func (mk *MockStore) UpdateSubOffset(projectUUID string, name string, offset int64) {

}

// ModSubPush modifies the subscription ack
func (mk *MockStore) ModSubPush(projectUUID string, name string, push string, rPolicy string, rPeriod int) error {
	return nil
}

// UpdateSubOffsetAck updates the offset of the current subscription
func (mk *MockStore) UpdateSubOffsetAck(projectUUID string, name string, offset int64, ts string) error {
	// find sub
	sub := QSub{}

	for _, item := range mk.SubList {
		if item.ProjectUUID == projectUUID && item.Name == name {
			sub = item
		}
	}

	// check if no ack pending
	if sub.NextOffset == 0 {
		return errors.New("no ack pending")
	}

	// check if ack offset is wrong - wrong ack
	if offset < sub.Offset || offset > sub.NextOffset {
		return errors.New("wrong ack")
	}

	// check if ack has timeout
	zSec := "2006-01-02T15:04:05Z"
	timeGiven, _ := time.Parse(zSec, ts)
	timeRef, _ := time.Parse(zSec, sub.PendingAck)
	durSec := timeGiven.Sub(timeRef).Seconds()

	if int(durSec) > sub.Ack {
		return errors.New("ack timeout")
	}

	return nil
}

// QueryProjects function queries for a specific project or for a list of all projects
func (mk *MockStore) QueryProjects(uuid string, name string) ([]QProject, error) {

	result := []QProject{}
	if name == "" && uuid == "" {
		result = mk.ProjectList
	} else if name != "" {
		for _, item := range mk.ProjectList {
			if item.Name == name {
				result = append(result, item)
				break
			}
		}
	} else if uuid != "" {
		for _, item := range mk.ProjectList {
			if item.UUID == uuid {
				result = append(result, item)
				break
			}
		}
	}

	if len(result) > 0 {
		return result, nil
	}

	return result, errors.New("not found")

}

// QueryUsers queries the datastore for user information
func (mk *MockStore) QueryUsers(projectUUID string, uuid string, name string) ([]QUser, error) {
	result := []QUser{}

	if name == "" && uuid == "" && projectUUID == "" {
		for _, item := range mk.UserList {
			result = append(result, item)
		}
	} else if name == "" && uuid == "" && projectUUID != "" {
		for _, item := range mk.UserList {
			if item.isInProject(projectUUID) {
				result = append(result, item)
			}
		}
	} else if uuid != "" {
		for _, item := range mk.UserList {
			if item.UUID == uuid {
				result = append(result, item)
			}
		}
	} else if name != "" {
		for _, item := range mk.UserList {
			if item.Name == name {
				result = append(result, item)
				break
			}
		}
	}

	if len(result) > 0 {
		return result, nil
	}

	return result, errors.New("not found")

}

// UpdateSubPull updates next offset info after a pull
func (mk *MockStore) UpdateSubPull(name string, offset int64, ts string) {

}

// Initialize is used to initalize the mock
func (mk *MockStore) Initialize() {

	// populate topics
	qtop1 := QTopic{"argo_uuid", "topic1"}
	qtop2 := QTopic{"argo_uuid", "topic2"}
	qtop3 := QTopic{"argo_uuid", "topic3"}
	mk.TopicList = append(mk.TopicList, qtop1)
	mk.TopicList = append(mk.TopicList, qtop2)
	mk.TopicList = append(mk.TopicList, qtop3)

	// populate Subscriptions
	qsub1 := QSub{"argo_uuid", "sub1", "topic1", 0, 0, "", "", 10, "linear", 300}
	qsub2 := QSub{"argo_uuid", "sub2", "topic2", 0, 0, "", "", 10, "linear", 300}
	qsub3 := QSub{"argo_uuid", "sub3", "topic3", 0, 0, "", "", 10, "linear", 300}
	qsub4 := QSub{"argo_uuid", "sub4", "topic4", 0, 0, "", "endpoint.foo", 10, "linear", 300}
	mk.SubList = append(mk.SubList, qsub1)
	mk.SubList = append(mk.SubList, qsub2)
	mk.SubList = append(mk.SubList, qsub3)
	mk.SubList = append(mk.SubList, qsub4)

	// populate Projects
	created := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	modified := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	qPr := QProject{UUID: "argo_uuid", Name: "ARGO", CreatedOn: created, ModifiedOn: modified, CreatedBy: "userA", Description: "simple project"}
	qPr2 := QProject{UUID: "argo_uuid2", Name: "ARGO2", CreatedOn: created, ModifiedOn: modified, CreatedBy: "userA", Description: "simple project"}
	mk.ProjectList = append(mk.ProjectList, qPr)
	mk.ProjectList = append(mk.ProjectList, qPr2)

	// populate Users
	qRole := []QProjectRoles{QProjectRoles{"argo_uuid", []string{"admin", "member"}}}
	qUsr := QUser{"uuid0", qRole, "Test", "S3CR3T", "Test@test.com", []string{}}
	mk.UserList = append(mk.UserList, qUsr)

	qRoleConsumer := []QProjectRoles{QProjectRoles{"argo_uuid", []string{"consumer"}}}
	qRoleProducer := []QProjectRoles{QProjectRoles{"argo_uuid", []string{"producer"}}}

	mk.UserList = append(mk.UserList, QUser{"uuid1", qRole, "UserA", "S3CR3T1", "foo-email", []string{}})
	mk.UserList = append(mk.UserList, QUser{"uuid2", qRole, "UserB", "S3CR3T2", "foo-email", []string{}})
	mk.UserList = append(mk.UserList, QUser{"uuid3", qRoleConsumer, "UserX", "S3CR3T3", "foo-email", []string{}})
	mk.UserList = append(mk.UserList, QUser{"uuid4", qRoleProducer, "UserZ", "S3CR3T4", "foo-email", []string{}})

	qRole1 := QRole{"topics:list_all", []string{"admin", "reader", "publisher"}}
	qRole2 := QRole{"topics:publish", []string{"admin", "publisher"}}
	mk.RoleList = append(mk.RoleList, qRole1)
	mk.RoleList = append(mk.RoleList, qRole2)

	qTopicACL01 := QAcl{[]string{"userA", "userB"}}
	qTopicACL02 := QAcl{[]string{"userA", "userB", "userD"}}
	qTopicACL03 := QAcl{[]string{"userC"}}

	qSubACL01 := QAcl{[]string{"userA", "userB"}}
	qSubACL02 := QAcl{[]string{"userA", "userC"}}
	qSubACL03 := QAcl{[]string{"userD", "userB", "userA"}}
	qSubACL04 := QAcl{[]string{"userB", "userD"}}

	mk.TopicsACL = make(map[string]QAcl)
	mk.SubsACL = make(map[string]QAcl)

	mk.TopicsACL["topic1"] = qTopicACL01
	mk.TopicsACL["topic2"] = qTopicACL02
	mk.TopicsACL["topic3"] = qTopicACL03

	mk.SubsACL["sub1"] = qSubACL01
	mk.SubsACL["sub2"] = qSubACL02
	mk.SubsACL["sub3"] = qSubACL03
	mk.SubsACL["sub4"] = qSubACL04

}

// QueryOneSub returns one sub exactly
func (mk *MockStore) QueryOneSub(projectUUID string, name string) (QSub, error) {
	for _, item := range mk.SubList {
		if item.Name == name && item.ProjectUUID == projectUUID {
			return item, nil
		}
	}

	return QSub{}, errors.New("empty")
}

// Clone the store
func (mk *MockStore) Clone() Store {
	return mk
}

// GetUserRoles returns the roles of a user in a project
func (mk *MockStore) GetUserRoles(projectUUID string, token string) ([]string, string) {
	for _, item := range mk.UserList {

		if item.Token == token {
			return item.getProjectRoles(projectUUID), item.Name

		}
	}

	return []string{}, ""
}

//HasResourceRoles returns the roles of a user in a project
func (mk *MockStore) HasResourceRoles(resource string, roles []string) bool {

	for _, item := range mk.RoleList {
		if item.Name == resource {
			for _, subitem := range item.Roles {
				for _, roleItem := range roles {
					if roleItem == subitem {
						return true
					}
				}
			}
		}

	}

	return false

}

// HasProject returns true if project exists in store
func (mk *MockStore) HasProject(name string) bool {
	for _, item := range mk.ProjectList {
		if item.Name == name {
			return true
		}
	}

	return false
}

// InsertTopic inserts a new topic object to the store
func (mk *MockStore) InsertTopic(projectUUID string, name string) error {
	topic := QTopic{ProjectUUID: projectUUID, Name: name}
	mk.TopicList = append(mk.TopicList, topic)
	return nil
}

// InsertSub inserts a new sub object to the store
func (mk *MockStore) InsertSub(projectUUID string, name string, topic string, offset int64, ack int, push string, rPolicy string, rPeriod int) error {
	sub := QSub{projectUUID, name, topic, offset, 0, "", push, ack, rPolicy, rPeriod}
	mk.SubList = append(mk.SubList, sub)
	return nil
}

// InsertProject inserts a project to the store
func (mk *MockStore) InsertProject(uuid string, name string, createdOn time.Time, modifiedOn time.Time, createdBy string, description string) error {
	project := QProject{UUID: uuid, Name: name, CreatedOn: createdOn, ModifiedOn: modifiedOn, CreatedBy: createdBy, Description: description}
	mk.ProjectList = append(mk.ProjectList, project)
	return nil
}

// RemoveProject removes an existing project
func (mk *MockStore) RemoveProject(uuid string) error {
	for i, project := range mk.ProjectList {
		if project.UUID == uuid {
			// found item at i, remove it using index
			mk.ProjectList = append(mk.ProjectList[:i], mk.ProjectList[i+1:]...)
			return nil
		}
	}

	return errors.New("not found")
}

// RemoveTopic removes an existing topic
func (mk *MockStore) RemoveTopic(projectUUID string, name string) error {
	for i, topic := range mk.TopicList {
		if topic.Name == name && topic.ProjectUUID == projectUUID {
			// found item at i, remove it using index
			mk.TopicList = append(mk.TopicList[:i], mk.TopicList[i+1:]...)
			return nil
		}
	}

	return errors.New("not found")
}

// RemoveProjectTopics removes all topics belonging to a specific project uuid
func (mk *MockStore) RemoveProjectTopics(projectUUID string) error {
	found := false
	newList := []QTopic{}
	for _, topic := range mk.TopicList {
		if topic.ProjectUUID != projectUUID {
			// found item at i, remove it using index
			newList = append(newList, topic)
		} else {
			found = true
		}
	}
	mk.TopicList = newList
	if found {
		return nil
	}
	return errors.New("not found")
}

// RemoveProjectSubs removes all existing subs belonging to a specific project uuid
func (mk *MockStore) RemoveProjectSubs(projectUUID string) error {
	found := false
	newList := []QSub{}
	for _, sub := range mk.SubList {
		if sub.ProjectUUID != projectUUID {
			// found item at i, remove it using index
			newList = append(newList, sub)
		} else {
			found = true
		}
	}
	mk.SubList = newList
	if found {
		return nil
	}
	return errors.New("not found")
}

// RemoveSub removes an existing sub from the store
func (mk *MockStore) RemoveSub(projectUUID string, name string) error {
	for i, sub := range mk.SubList {
		if sub.Name == name && sub.ProjectUUID == projectUUID {
			// found item at i, remove it using index
			mk.SubList = append(mk.SubList[:i], mk.SubList[i+1:]...)
			return nil
		}
	}

	return errors.New("not found")
}

// QueryPushSubs Query push Subscription info from store
func (mk *MockStore) QueryPushSubs() []QSub {
	return mk.SubList
}

// QuerySubs Query Subscription info from store
func (mk *MockStore) QuerySubs(projectUUID string, name string) ([]QSub, error) {
	result := []QSub{}
	for _, item := range mk.SubList {
		if projectUUID == item.ProjectUUID {
			if name == "" {
				result = append(result, item)
			} else if name == item.Name {
				return []QSub{item}, nil
			}
		}
	}

	return result, nil
}

// QueryTopics Query Subscription info from store
func (mk *MockStore) QueryTopics(projectUUID string, name string) ([]QTopic, error) {
	result := []QTopic{}
	for _, item := range mk.TopicList {
		if projectUUID == item.ProjectUUID {
			if name == "" {
				result = append(result, item)
			} else if name == item.Name {
				return []QTopic{item}, nil
			}
		}
	}

	return result, nil
}
