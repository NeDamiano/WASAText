package database

// CreateGroup creates a group with the given name and members
func (db *appdbimpl) CreateGroup(name string, members []int64) (int64, error) {
	res, err := db.c.Exec("INSERT INTO conversations (id, is_group, name, photo) VALUES (NULL, 1, ?, NULL)", name)
	if err != nil {
		return 0, err
	}
	groupId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, userId := range members {
		err = db.AddMember(groupId, userId)
		if err != nil {
			return 0, err
		}
	}
	return groupId, nil
}

// AddMember adds a user to a group (nothing happens if they are already in it)
func (db *appdbimpl) AddMember(groupId int64, userId int64) error {
	_, err := db.c.Exec("INSERT OR IGNORE INTO members (user_id, conversation_id) VALUES (?, ?)", userId, groupId)
	return err
}

// LeaveGroup removes the user from the group, and their read status of its messages
func (db *appdbimpl) LeaveGroup(groupId int64, userId int64) error {
	_, err := db.c.Exec("DELETE FROM members WHERE user_id = ? AND conversation_id = ?", userId, groupId)
	if err != nil {
		return err
	}
	_, err = db.c.Exec(`DELETE FROM message_status WHERE user_id = ?
		AND message_id IN (SELECT id FROM messages WHERE conversation_id = ?)`, userId, groupId)
	return err
}

// SetGroupName changes the name of a group
func (db *appdbimpl) SetGroupName(groupId int64, name string) error {
	_, err := db.c.Exec("UPDATE conversations SET name = ? WHERE id = ?", name, groupId)
	return err
}

// SetGroupPhoto changes the photo URL of a group
func (db *appdbimpl) SetGroupPhoto(groupId int64, photo string) error {
	_, err := db.c.Exec("UPDATE conversations SET photo = ? WHERE id = ?", photo, groupId)
	return err
}
