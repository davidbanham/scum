package notifications

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/davidbanham/marcel"
	scumfilter "github.com/davidbanham/scum/filter"
	scumflashes "github.com/davidbanham/scum/flash"
	scummodel "github.com/davidbanham/scum/model"
	scumquery "github.com/davidbanham/scum/query"
	scumsearch "github.com/davidbanham/scum/search"

	"github.com/davidbanham/human_duration"
	kewpie "github.com/davidbanham/kewpie_go/v3"
	uuid "github.com/satori/go.uuid"
)

type Publisher interface {
	Publish(context.Context, *kewpie.Task) error
}

type Notification struct {
	Queue          Publisher
	ID             string
	Topic          string
	Level          scumflashes.FlashLevel
	Template       string
	Vars           NotificationVars
	Target         string
	OrganisationID string
	UserID         string
	UserEmail      string
	Emailed        bool
	EmailedAt      time.Time
	Seen           bool
	SeenAt         time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Revision       string
}

type NotificationVars []any

func (this NotificationVars) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *NotificationVars) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &this)
}

func (this *Notification) colmap() *scummodel.Colmap {
	return &scummodel.Colmap{
		"notifications.id":              &this.ID,
		"notifications.topic":           &this.Topic,
		"notifications.nlevel":          &this.Level,
		"notifications.template":        &this.Template,
		"notifications.vars":            &this.Vars,
		"notifications.target":          &this.Target,
		"notifications.organisation_id": &this.OrganisationID,
		"notifications.user_id":         &this.UserID,
		"users.email":                   &this.UserEmail,
		"notifications.emailed":         &this.Emailed,
		"notifications.emailed_at":      &this.EmailedAt,
		"notifications.seen":            &this.Seen,
		"notifications.seen_at":         &this.SeenAt,
		"notifications.created_at":      &this.CreatedAt,
		"notifications.updated_at":      &this.UpdatedAt,
		"notifications.revision":        &this.Revision,
	}
}

func (this *Notification) New(level scumflashes.FlashLevel, topic, organisationID, userID string) {
	this.ID = uuid.NewV4().String()
	this.Level = level
	this.Topic = topic
	this.OrganisationID = organisationID
	this.UserID = userID
	this.CreatedAt = time.Now()
	this.UpdatedAt = time.Now()
}

func (this *Notification) FindByColumn(ctx context.Context, col, val string) error {
	q, props := scummodel.FindByColumn("notifications JOIN users ON users.id = notifications.user_id", this.colmap(), col)
	return scummodel.ExecFindByColumn(ctx, q, val, props)
}

func (this *Notification) FindByID(ctx context.Context, id string) error {
	return this.FindByColumn(ctx, "notifications.id", id)
}

func (this Notification) Delete(ctx context.Context) error {
	db := ctx.Value("tx").(scummodel.Querier)

	_, err := db.ExecContext(ctx, this.auditQuery(ctx, "D")+"DELETE FROM notifications WHERE id = $1 AND revision = $2", this.ID, this.Revision)
	return err
}

func (this Notification) auditQuery(ctx context.Context, action string) string {
	return ""
}

func (this *Notification) Save(ctx context.Context) error {
	q, props, newRev := scummodel.StandardSave("notifications", this.colmap(), this.auditQuery(ctx, "U"))

	if err := scummodel.ExecSave(ctx, q, props); err != nil {
		return err
	}

	this.Revision = newRev

	if !this.Emailed && !this.Seen {
		task := kewpie.Task{}
		if err := task.Marshal(this); err != nil {
			return err
		}
		task.Delay = 10 * time.Minute

		queue := ctx.Value("queue").(kewpie.Kewpie)

		if err := queue.Publish(ctx, "unread_notification", &task); err != nil {
			return err
		}
	}

	return nil
}

func (this *Notification) Publish(ctx context.Context) error {
	task := kewpie.Task{}
	if err := task.Marshal(this); err != nil {
		return err
	}
	task.Delay = 10 * time.Minute

	if err := this.Queue.Publish(ctx, &task); err != nil {
		return err
	}
	return nil
}

func (this Notification) Label() string {
	return fmt.Sprintf(this.Template, this.Vars...)
}

func (this Notification) Ago() string {
	return human_duration.String(time.Now().Sub(this.CreatedAt), "minute") + " ago"
}

type Notifications struct {
	Data     []Notification
	Criteria scummodel.Criteria
}

func (this Notifications) colmap() *scummodel.Colmap {
	r := Notification{}
	return r.colmap()
}

func (Notifications) AvailableFilters() scumfilter.Filters {
	return scumfilter.CommonFilters("notifications")
}

func (Notifications) Searchable() scumsearch.Searchable {
	return scumsearch.Searchable{
		EntityType: "Notification",
		Label:      "topic",
		Path:       "notifications",
		Tablename:  "notifications",
		Permitted:  scumsearch.BasicRoleCheck("admin"),
	}
}

func (this Notifications) ByID() map[string]Notification {
	ret := map[string]Notification{}
	for _, t := range this.Data {
		ret[t.ID] = t
	}
	return ret
}

func (this Notifications) ByTopic() map[string]Notifications {
	ret := map[string]Notifications{}
	for _, n := range this.Data {
		nots := ret[n.Topic]
		nots.Data = append(nots.Data, n)
		ret[n.Topic] = nots
	}
	return ret
}

func (this Notifications) ForTopic(in string) Notifications {
	ret := Notifications{}
	for _, n := range this.Data {
		if n.Topic == in {
			ret.Data = append(ret.Data, n)
		}
	}
	return ret
}

func (this Notifications) Unseen() Notifications {
	ret := Notifications{}
	for _, n := range this.Data {
		if !n.Seen {
			ret.Data = append(ret.Data, n)
		}
	}
	return ret
}

func (this Notifications) Seen() Notifications {
	ret := Notifications{}
	for _, n := range this.Data {
		if n.Seen {
			ret.Data = append(ret.Data, n)
		}
	}
	return ret
}

func (this *Notifications) FindAll(ctx context.Context, criteria scummodel.Criteria) error {
	this.Criteria = criteria

	db := ctx.Value("tx").(scummodel.Querier)

	cols, _ := this.colmap().Split()

	var rows *sql.Rows
	var err error

	switch v := criteria.Query.(type) {
	default:
		return scummodel.ErrInvalidQuery{Query: v, Model: "notifications"}
	case scumquery.Query:
		rows, err = db.QueryContext(ctx, v.Construct(cols, "notifications JOIN users ON users.id = notifications.user_id", criteria.Filters, criteria.Pagination, scumquery.Order{Desc: true, By: "notifications.created_at"}), v.Args()...)
	}
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		notification := Notification{}
		props := notification.colmap().ByKeys(cols)
		if err := rows.Scan(props...); err != nil {
			return err
		}
		(*this).Data = append((*this).Data, notification)
	}
	return err
}

func (this *Notifications) UnseenAndNotEmailedForUser(ctx context.Context, userID string) error {
	criteria := scummodel.Criteria{
		Query: &scumquery.ByUser{ID: userID},
	}
	notEmailedFilter := scumfilter.HasProp{}
	notEmailedFilter.Hydrate(scumfilter.HasPropOpts{
		Label: "not-emailed",
		ID:    "not-emailed",
		Table: "notifications",
		Col:   "emailed",
		Value: "false",
	})
	notSeenFilter := scumfilter.HasProp{}
	notSeenFilter.Hydrate(scumfilter.HasPropOpts{
		Label: "not-seen",
		ID:    "not-seen",
		Table: "notifications",
		Col:   "seen",
		Value: "false",
	})
	unseenNotEmailedForUser := scumfilter.FilterSet{
		IsAnd:    true,
		Filters:  scumfilter.Filters{&notSeenFilter, &notEmailedFilter},
		CustomID: fmt.Sprintf("unseen-and-not-emailed-for-user"),
	}
	criteria.Filters = append(criteria.Filters, &unseenNotEmailedForUser)

	return this.FindAll(ctx, criteria)
}

func (this Notifications) EmailCopy(uri string) marcel.Email {
	email := marcel.Email{}

	text := []string{}
	html := []string{}

	for _, notification := range this.Data {
		text = append(text, fmt.Sprintf("%s\r\n\r\n%s", notification.Label(), uri+notification.Target))
		html = append(html, fmt.Sprintf(`<p>%s</p><p><a href="%s">Click here to view.</a></p>`, notification.Label(), uri+notification.Target))

		email.Subject = notification.Label()
	}

	email.Text = strings.Join(text, "\r\n\r\n")
	email.HTML = strings.Join(html, "<br><br>")

	if len(this.Data) > 1 {
		email.Subject = "You have multiple unread notifications"
	}

	return email
}

func (this *Notifications) MarkEmailed(ctx context.Context) error {
	for _, notification := range this.Data {
		notification.Emailed = true
		notification.EmailedAt = time.Now()
		if err := notification.Save(ctx); err != nil {
			return err
		}
	}

	return nil
}

type NotificationStatus struct {
	NoUnread    bool
	Num         int
	NumWarnings int
	NumSuccess  int
	NumInfo     int
	AnyWarnings bool
	AnySuccess  bool
	AnyInfo     bool
}

func NotificationStatusForUser(ctx context.Context, userID string) (NotificationStatus, error) {
	db := ctx.Value("tx").(scummodel.Querier)
	var warn, success, info int
	err := db.QueryRowContext(ctx, `SELECT 
	COUNT(*) FILTER(WHERE nlevel = '1') AS warnings,
	COUNT(*) FILTER(WHERE nlevel = '2') AS success,
	COUNT(*) FILTER(WHERE nlevel = '3') AS info
	FROM notifications WHERE user_id = $1 AND seen = false`, userID).Scan(&warn, &success, &info)
	stats := NotificationStatus{
		NoUnread:    warn+success+info == 0,
		Num:         warn + success + info,
		NumWarnings: warn,
		NumSuccess:  success,
		NumInfo:     info,
		AnyWarnings: warn > 0,
		AnySuccess:  success > 0,
		AnyInfo:     info > 0,
	}
	return stats, err
}

func NotificationStatusForUserInOrg(ctx context.Context, userID, orgID string) (NotificationStatus, error) {
	db := ctx.Value("tx").(scummodel.Querier)
	var warn, success, info int
	err := db.QueryRowContext(ctx, `SELECT 
	COUNT(*) FILTER(WHERE nlevel = '1') AS warnings,
	COUNT(*) FILTER(WHERE nlevel = '2') AS success,
	COUNT(*) FILTER(WHERE nlevel = '3') AS info
	FROM notifications WHERE user_id = $1 AND organisation_id = $2 AND seen = false`, userID, orgID).Scan(&warn, &success, &info)
	stats := NotificationStatus{
		NoUnread:    warn+success+info == 0,
		Num:         warn + success + info,
		NumWarnings: warn,
		NumSuccess:  success,
		NumInfo:     info,
		AnyWarnings: warn > 0,
		AnySuccess:  success > 0,
		AnyInfo:     info > 0,
	}
	return stats, err
}
