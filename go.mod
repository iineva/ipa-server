module github.com/iineva/ipa-server

go 1.16

require (
	github.com/go-kit/kit v0.10.0
	github.com/google/uuid v1.2.0
	github.com/lithammer/shortuuid v3.0.0+incompatible
	github.com/poolqa/CgbiPngFix v0.0.0-20200429152610-b5884815004a
	github.com/spf13/afero v1.6.0
	howett.net/plist v0.0.0-20201203080718-1454fab16a06
)

replace github.com/poolqa/CgbiPngFix => github.com/iineva/CgbiPngFix v0.0.0-20210522155758-504358a2f2de
