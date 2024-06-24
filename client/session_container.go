package client

// // sessionContainer is a special type which unifies session logic management for client parameters.
// // All methods make public, because sessionContainer is included in Prm* structs.
// type sessionContainer struct {
// 	isSessionIgnored bool
// 	meta             v2session.RequestMetaHeader
// }
//
// // GetSession returns session object.
// //
// // Returns:
// //   - [ErrNoSession] err if session wasn't set.
// //   - [ErrNoSessionExplicitly] if IgnoreSession was used.
// func (x *sessionContainer) GetSession() (*session.Object, error) {
// 	if x.isSessionIgnored {
// 		return nil, ErrNoSessionExplicitly
// 	}
//
// 	token := x.meta.GetSessionToken()
// 	if token == nil {
// 		return nil, ErrNoSession
// 	}
//
// 	var sess session.Object
// 	if err := sess.ReadFromV2(*token); err != nil {
// 		return nil, err
// 	}
//
// 	return &sess, nil
// }
//
// // WithinSession specifies session within which the query must be executed.
// //
// // Creator of the session acquires the authorship of the request.
// // This may affect the execution of an operation (e.g. access control).
// //
// // See also IgnoreSession.
// //
// // Must be signed.
// func (x *sessionContainer) WithinSession(t session.Object) {
// 	var tokv2 v2session.Token
// 	t.WriteToV2(&tokv2)
// 	x.meta.SetSessionToken(&tokv2)
// 	x.isSessionIgnored = false
// }
//
// // IgnoreSession disables auto-session creation.
// //
// // See also WithinSession.
// func (x *sessionContainer) IgnoreSession() {
// 	x.isSessionIgnored = true
// 	x.meta.SetSessionToken(nil)
// }
