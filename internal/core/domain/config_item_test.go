package domain

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAddConfigItem(t *testing.T) {
	t.Run("Test add item to set", func(t *testing.T) {
		set := NewConfigSet("test")
		item := ConfigItem{
			Key:   "itemKey",
			Value: 10,
			Type:  Plain,
		}

		err := set.Add(item)

		if err != nil {
			t.Errorf("Expected add without error, got: %v", err)
		}

		if !cmp.Equal(set.Items[item.Key], item) {
			t.Errorf("Expected %v to be: %v", set.Items[item.Key], item)
		}
	})

	t.Run("Test add item duplicated to set", func(t *testing.T) {
		set := NewConfigSet("test")
		item := ConfigItem{
			Key:   "itemKey",
			Value: 10,
			Type:  Plain,
		}

		err := set.Add(item)
		if err != nil {
			t.Errorf("Expected add without error, got: %v", err)
		}

		otherItem := ConfigItem{
			Key:   "itemKey",
			Value: 100,
			Type:  Plain,
		}

		err = set.Add(otherItem)
		if err == nil {
			t.Errorf("Expected error: %v, got nil", ErrDuplicatedKey)
		}

		if !cmp.Equal(set.Items[item.Key], item) {
			t.Errorf("Expected %v to be: %v", set.Items[item.Key], item)
		}
	})
}

func TestGetConfigItem(t *testing.T) {
	t.Run("Test get item from set", func(t *testing.T) {
		set := NewConfigSet("test")
		item := ConfigItem{
			Key:   "itemKey",
			Value: 10,
			Type:  Plain,
		}

		set.Add(item)
		got, err := set.Get("itemKey")

		if err != nil {
			t.Errorf("Expected get without error, got: %v", err)
		}

		if !cmp.Equal(got, item) {
			t.Errorf("Expected %v to be: %v", set.Items[item.Key], item)
		}
	})

	t.Run("Test get nonexisting item", func(t *testing.T) {
		set := NewConfigSet("test")

		_, err := set.Get("itemkey")

		if err == nil {
			t.Errorf("Expected error: %v got nil", ErrKeyNotExists)
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Run("Test update item from set", func(t *testing.T) {
		set := NewConfigSet("test")
		item := ConfigItem{
			Key:   "itemKey",
			Value: 10,
			Type:  Plain,
		}

		err := set.Add(item)
		if err != nil {
			t.Errorf("Expected add without error, got: %v", err)
		}

		otherItem := ConfigItem{
			Key:   "itemKey",
			Value: 100,
			Type:  Secret,
		}

		got, err := set.Update(otherItem)
		if err != nil {
			t.Errorf("Expected update without error, got: %v", err)
		}

		if !cmp.Equal(got, otherItem) {
			t.Errorf("Expected %v to be: %v", otherItem, got)
		}

		if !cmp.Equal(set.Items[item.Key], otherItem) {
			t.Errorf("Expected %v to be: %v", set.Items[item.Key], otherItem)
		}
	})

	t.Run("Test update nonexisting item", func(t *testing.T) {
		set := NewConfigSet("test")

		item := ConfigItem{
			Key:   "itemKey",
			Value: 10,
			Type:  Plain,
		}
		_, err := set.Update(item)

		if err == nil {
			t.Errorf("Expected error: %v got nil", ErrKeyNotExists)
		}
	})
}

func TestDelete(t *testing.T) {
	t.Run("Test delete item from set", func(t *testing.T) {
		set := NewConfigSet("test")
		item := ConfigItem{
			Key:   "itemKey",
			Value: 10,
			Type:  Plain,
		}

		set.Add(item)
		got, err := set.Delete("itemKey")

		if err != nil {
			t.Errorf("Expected delete without error, got: %v", err)
		}

		if !cmp.Equal(got, item) {
			t.Errorf("Expected %v to be: %v", set.Items[item.Key], item)
		}

		_, err = set.Get("itemKey")

		if err == nil {
			t.Errorf("Expected error: %v, got nil", ErrKeyNotExists)
		}
	})

	t.Run("Test delete nonexisting item", func(t *testing.T) {
		set := NewConfigSet("test")

		_, err := set.Delete("itemkey")

		if err == nil {
			t.Errorf("Expected error: %v got nil", ErrKeyNotExists)
		}
	})
}
