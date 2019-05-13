// Copyright (c) 2017 Shivaram Lingamneni <slingamn@cs.stanford.edu>
// released under the MIT license

package irc

import (
	"net"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/oragono/oragono/irc/languages"
	"github.com/oragono/oragono/irc/modes"
)

func (server *Server) Config() (config *Config) {
	return (*Config)(atomic.LoadPointer(&server.config))
}

func (server *Server) SetConfig(config *Config) {
	atomic.StorePointer(&server.config, unsafe.Pointer(config))
}

func (server *Server) Limits() Limits {
	return server.Config().Limits
}

func (server *Server) Password() []byte {
	return server.Config().Server.passwordBytes
}

func (server *Server) RecoverFromErrors() bool {
	return *server.Config().Debug.RecoverFromErrors
}

func (server *Server) DefaultChannelModes() modes.Modes {
	return server.Config().Channels.defaultModes
}

func (server *Server) ChannelRegistrationEnabled() bool {
	return server.Config().Channels.Registration.Enabled
}

func (server *Server) AccountConfig() *AccountConfig {
	return &server.Config().Accounts
}

func (server *Server) GetOperator(name string) (oper *Oper) {
	name, err := CasefoldName(name)
	if err != nil {
		return
	}
	return server.Config().operators[name]
}

func (server *Server) Languages() (lm *languages.Manager) {
	return server.Config().languageManager
}

func (client *Client) Sessions() (sessions []*Session) {
	client.stateMutex.RLock()
	sessions = make([]*Session, len(client.sessions))
	copy(sessions, client.sessions)
	client.stateMutex.RUnlock()
	return
}

type SessionData struct {
	ctime    time.Time
	atime    time.Time
	ip       net.IP
	hostname string
}

func (client *Client) AllSessionData(currentSession *Session) (data []SessionData, currentIndex int) {
	currentIndex = -1
	client.stateMutex.RLock()
	defer client.stateMutex.RUnlock()

	data = make([]SessionData, len(client.sessions))
	for i, session := range client.sessions {
		if session == currentSession {
			currentIndex = i
		}
		data[i] = SessionData{
			atime:    session.atime,
			ctime:    session.ctime,
			hostname: session.rawHostname,
		}
		if session.proxiedIP != nil {
			data[i].ip = session.proxiedIP
		} else {
			data[i].ip = session.realIP
		}
	}
	return
}

func (client *Client) AddSession(session *Session) (success bool) {
	client.stateMutex.Lock()
	defer client.stateMutex.Unlock()

	if len(client.sessions) == 0 {
		return false
	}
	session.client = client
	client.sessions = append(client.sessions, session)
	return true
}

func (client *Client) removeSession(session *Session) (success bool, length int) {
	if len(client.sessions) == 0 {
		return
	}
	sessions := make([]*Session, 0, len(client.sessions)-1)
	for _, currentSession := range client.sessions {
		if session == currentSession {
			success = true
		} else {
			sessions = append(sessions, currentSession)
		}
	}
	client.sessions = sessions
	length = len(sessions)
	return
}

func (client *Client) Nick() string {
	client.stateMutex.RLock()
	defer client.stateMutex.RUnlock()
	return client.nick
}

func (client *Client) NickMaskString() string {
	client.stateMutex.RLock()
	defer client.stateMutex.RUnlock()
	return client.nickMaskString
}

func (client *Client) NickCasefolded() string {
	client.stateMutex.RLock()
	defer client.stateMutex.RUnlock()
	return client.nickCasefolded
}

func (client *Client) NickMaskCasefolded() string {
	client.stateMutex.RLock()
	defer client.stateMutex.RUnlock()
	return client.nickMaskCasefolded
}

func (client *Client) Username() string {
	client.stateMutex.RLock()
	defer client.stateMutex.RUnlock()
	return client.username
}

func (client *Client) Hostname() string {
	client.stateMutex.RLock()
	defer client.stateMutex.RUnlock()
	return client.hostname
}

func (client *Client) Realname() string {
	client.stateMutex.RLock()
	defer client.stateMutex.RUnlock()
	return client.realname
}

func (client *Client) Away() (result bool) {
	client.stateMutex.Lock()
	result = client.away
	client.stateMutex.Unlock()
	return
}

func (client *Client) SetAway(away bool, awayMessage string) (changed bool) {
	client.stateMutex.Lock()
	changed = away != client.away
	client.away = away
	client.awayMessage = awayMessage
	client.stateMutex.Unlock()
	return
}

// uniqueIdentifiers returns the strings for which the server enforces per-client
// uniqueness/ownership; no two clients can have colliding casefolded nicks or
// skeletons.
func (client *Client) uniqueIdentifiers() (nickCasefolded string, skeleton string) {
	client.stateMutex.RLock()
	defer client.stateMutex.RUnlock()
	return client.nickCasefolded, client.skeleton
}

func (client *Client) ResumeID() string {
	client.stateMutex.RLock()
	defer client.stateMutex.RUnlock()
	return client.resumeID
}

func (client *Client) SetResumeID(id string) {
	client.stateMutex.Lock()
	defer client.stateMutex.Unlock()
	client.resumeID = id
}

func (client *Client) Oper() *Oper {
	client.stateMutex.RLock()
	defer client.stateMutex.RUnlock()
	return client.oper
}

func (client *Client) Registered() bool {
	client.stateMutex.RLock()
	defer client.stateMutex.RUnlock()
	return client.registered
}

func (client *Client) SetRegistered() {
	// `registered` is only written from the client's own goroutine, but may be
	// read from other goroutines; therefore, the client's own goroutine may read
	// the value without synchronization, but must write it with synchronization,
	// and other goroutines must read it with synchronization
	client.stateMutex.Lock()
	client.registered = true
	client.stateMutex.Unlock()
}

func (client *Client) RawHostname() (result string) {
	client.stateMutex.Lock()
	result = client.rawHostname
	client.stateMutex.Unlock()
	return
}

func (client *Client) AwayMessage() (result string) {
	client.stateMutex.RLock()
	result = client.awayMessage
	client.stateMutex.RUnlock()
	return
}

func (client *Client) SetAwayMessage(message string) {
	client.stateMutex.Lock()
	client.awayMessage = message
	client.stateMutex.Unlock()
}

func (client *Client) Account() string {
	client.stateMutex.RLock()
	defer client.stateMutex.RUnlock()
	return client.account
}

func (client *Client) AccountName() string {
	client.stateMutex.RLock()
	defer client.stateMutex.RUnlock()
	return client.accountName
}

func (client *Client) SetAccountName(account string) (changed bool) {
	var casefoldedAccount string
	var err error
	if account != "" {
		if casefoldedAccount, err = CasefoldName(account); err != nil {
			return
		}
	}

	client.stateMutex.Lock()
	defer client.stateMutex.Unlock()
	changed = client.account != casefoldedAccount
	client.account = casefoldedAccount
	client.accountName = account
	return
}

func (client *Client) Languages() (languages []string) {
	client.stateMutex.RLock()
	languages = client.languages
	client.stateMutex.RUnlock()
	return languages
}

func (client *Client) SetLanguages(languages []string) {
	client.stateMutex.Lock()
	client.languages = languages
	client.stateMutex.Unlock()
}

func (client *Client) HasMode(mode modes.Mode) bool {
	// client.flags has its own synch
	return client.flags.HasMode(mode)
}

func (client *Client) SetMode(mode modes.Mode, on bool) bool {
	return client.flags.SetMode(mode, on)
}

func (client *Client) Channels() (result []*Channel) {
	client.stateMutex.RLock()
	defer client.stateMutex.RUnlock()
	length := len(client.channels)
	result = make([]*Channel, length)
	i := 0
	for channel := range client.channels {
		result[i] = channel
		i++
	}
	return
}

func (client *Client) NumChannels() int {
	client.stateMutex.RLock()
	defer client.stateMutex.RUnlock()
	return len(client.channels)
}

func (client *Client) WhoWas() (result WhoWas) {
	return client.Details().WhoWas
}

func (client *Client) Details() (result ClientDetails) {
	client.stateMutex.RLock()
	defer client.stateMutex.RUnlock()
	return client.detailsNoMutex()
}

func (client *Client) detailsNoMutex() (result ClientDetails) {
	result.nick = client.nick
	result.nickCasefolded = client.nickCasefolded
	result.username = client.username
	result.hostname = client.hostname
	result.realname = client.realname
	result.nickMask = client.nickMaskString
	result.nickMaskCasefolded = client.nickMaskCasefolded
	result.account = client.account
	result.accountName = client.accountName
	return
}

func (channel *Channel) Name() string {
	channel.stateMutex.RLock()
	defer channel.stateMutex.RUnlock()
	return channel.name
}

func (channel *Channel) NameCasefolded() string {
	channel.stateMutex.RLock()
	defer channel.stateMutex.RUnlock()
	return channel.nameCasefolded
}

func (channel *Channel) Rename(name, nameCasefolded string) {
	channel.stateMutex.Lock()
	channel.name = name
	channel.nameCasefolded = nameCasefolded
	if channel.registeredFounder != "" {
		channel.registeredTime = time.Now().UTC()
	}
	channel.stateMutex.Unlock()
}

func (channel *Channel) Members() (result []*Client) {
	channel.stateMutex.RLock()
	defer channel.stateMutex.RUnlock()
	return channel.membersCache
}

func (channel *Channel) setUserLimit(limit int) {
	channel.stateMutex.Lock()
	channel.userLimit = limit
	channel.stateMutex.Unlock()
}

func (channel *Channel) Key() string {
	channel.stateMutex.RLock()
	defer channel.stateMutex.RUnlock()
	return channel.key
}

func (channel *Channel) setKey(key string) {
	channel.stateMutex.Lock()
	defer channel.stateMutex.Unlock()
	channel.key = key
}

func (channel *Channel) Founder() string {
	channel.stateMutex.RLock()
	defer channel.stateMutex.RUnlock()
	return channel.registeredFounder
}

func (channel *Channel) DirtyBits() (dirtyBits uint) {
	channel.stateMutex.Lock()
	dirtyBits = channel.dirtyBits
	channel.stateMutex.Unlock()
	return
}

func (channel *Channel) HighestUserMode(client *Client) (result modes.Mode) {
	channel.stateMutex.RLock()
	clientModes := channel.members[client]
	channel.stateMutex.RUnlock()
	return clientModes.HighestChannelUserMode()
}
