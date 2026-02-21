package acl

type StorageAclType int

const (
	AclTypePrivate StorageAclType = iota
	AclTypePublicRead
)
