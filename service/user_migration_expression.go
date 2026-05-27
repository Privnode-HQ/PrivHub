package service

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

type migrationExprTokenType int

const (
	migrationExprTokenEOF migrationExprTokenType = iota
	migrationExprTokenIdent
	migrationExprTokenString
	migrationExprTokenNumber
	migrationExprTokenOp
	migrationExprTokenAnd
	migrationExprTokenOr
	migrationExprTokenNot
	migrationExprTokenLParen
	migrationExprTokenRParen
	migrationExprTokenLBrace
	migrationExprTokenRBrace
	migrationExprTokenComma
)

type migrationExprToken struct {
	typ migrationExprTokenType
	val string
}

type migrationExprNode interface {
	eval(user model.User) (bool, error)
}

type migrationExprCompareNode struct {
	field  string
	op     string
	values []string
}

type migrationExprBinaryNode struct {
	op          string
	left, right migrationExprNode
}

type migrationExprNotNode struct {
	child migrationExprNode
}

type migrationExprParser struct {
	tokens []migrationExprToken
	pos    int
}

type UserMigrationExpressionDocField struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Operators   []string `json:"operators"`
	Examples    []string `json:"examples"`
}

func UserMigrationExpressionDocs() map[string]any {
	return map[string]any{
		"syntax": []string{
			`字段 运算符 值，例如：email matches ".*@example\\.com$"`,
			`使用 and、or、not 和括号组合条件，例如：(group == "vip" or role == "admin") and status == "enabled"`,
			`in 使用集合，例如：provider in {"github", "oidc"}`,
			`字符串可以使用双引号或单引号；未加引号的 enabled、admin、github 等会按普通字符串处理。`,
		},
		"operators": []string{"==", "!=", "eq", "ne", "contains", "matches", "in"},
		"fields": []UserMigrationExpressionDocField{
			{
				Name:        "username",
				Description: "用户名。",
				Operators:   []string{"==", "!=", "contains", "matches", "in"},
				Examples:    []string{`username matches "^team-"`, `username contains "test"`},
			},
			{
				Name:        "email",
				Description: "用户当前邮箱。",
				Operators:   []string{"==", "!=", "contains", "matches", "in"},
				Examples:    []string{`email matches ".*@example\\.com$"`},
			},
			{
				Name:        "cah",
				Description: "Customer Account Handle，可写 cah 或 cah_id。",
				Operators:   []string{"==", "!=", "matches", "in"},
				Examples:    []string{`cah == "ABCDE1"`},
			},
			{
				Name:        "provider",
				Description: "第三方登录方式，支持 github、discord、oidc、wechat、telegram、linuxdo、password。",
				Operators:   []string{"==", "!=", "matches", "in"},
				Examples:    []string{`provider in {"github", "oidc"}`},
			},
			{
				Name:        "status",
				Description: "用户状态，可用 enabled、disabled 或数字 1、2。",
				Operators:   []string{"==", "!=", "in"},
				Examples:    []string{`status == "enabled"`},
			},
			{
				Name:        "role",
				Description: "角色，可用 guest、common、support、admin、root 或数字 0、1、5、10、100。",
				Operators:   []string{"==", "!=", "in"},
				Examples:    []string{`role in {"common", "support"}`},
			},
			{
				Name:        "group",
				Description: "用户分组。",
				Operators:   []string{"==", "!=", "contains", "matches", "in"},
				Examples:    []string{`group == "vip"`},
			},
		},
		"examples": []string{
			`status == "enabled" and email matches ".*@example\\.com$"`,
			`provider in {"github", "oidc"} and group == "paid"`,
			`(role == "common" or role == "support") and not email contains "+test"`,
		},
	}
}

func ValidateUserMigrationExpression(expression string) error {
	_, err := parseUserMigrationExpression(expression)
	return err
}

func MatchUserMigrationExpression(expression string, user model.User) (bool, error) {
	node, err := parseUserMigrationExpression(expression)
	if err != nil {
		return false, err
	}
	return node.eval(user)
}

func MigrationExpressionMentionsField(expression string, field string) bool {
	tokens, err := tokenizeMigrationExpression(expression)
	if err != nil {
		return false
	}
	field = normalizeMigrationExpressionField(field)
	for _, token := range tokens {
		if token.typ == migrationExprTokenIdent && normalizeMigrationExpressionField(token.val) == field {
			return true
		}
	}
	return false
}

func parseUserMigrationExpression(expression string) (migrationExprNode, error) {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return nil, fmt.Errorf("迁移筛选表达式不能为空")
	}
	tokens, err := tokenizeMigrationExpression(expression)
	if err != nil {
		return nil, err
	}
	parser := &migrationExprParser{tokens: tokens}
	node, err := parser.parseOr()
	if err != nil {
		return nil, err
	}
	if parser.peek().typ != migrationExprTokenEOF {
		return nil, fmt.Errorf("表达式包含无法解析的内容：%s", parser.peek().val)
	}
	return node, nil
}

func tokenizeMigrationExpression(input string) ([]migrationExprToken, error) {
	tokens := make([]migrationExprToken, 0)
	for i := 0; i < len(input); {
		r := rune(input[i])
		if unicode.IsSpace(r) {
			i++
			continue
		}
		switch input[i] {
		case '(':
			tokens = append(tokens, migrationExprToken{typ: migrationExprTokenLParen, val: "("})
			i++
		case ')':
			tokens = append(tokens, migrationExprToken{typ: migrationExprTokenRParen, val: ")"})
			i++
		case '{':
			tokens = append(tokens, migrationExprToken{typ: migrationExprTokenLBrace, val: "{"})
			i++
		case '}':
			tokens = append(tokens, migrationExprToken{typ: migrationExprTokenRBrace, val: "}"})
			i++
		case ',':
			tokens = append(tokens, migrationExprToken{typ: migrationExprTokenComma, val: ","})
			i++
		case '=', '!':
			if i+1 < len(input) && input[i+1] == '=' {
				tokens = append(tokens, migrationExprToken{typ: migrationExprTokenOp, val: input[i : i+2]})
				i += 2
				continue
			}
			return nil, fmt.Errorf("无效运算符：%s", input[i:i+1])
		case '"', '\'':
			value, next, err := readMigrationExprString(input, i)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, migrationExprToken{typ: migrationExprTokenString, val: value})
			i = next
		default:
			if isMigrationExprIdentStart(r) {
				start := i
				i++
				for i < len(input) && isMigrationExprIdentPart(rune(input[i])) {
					i++
				}
				value := input[start:i]
				switch strings.ToLower(value) {
				case "and":
					tokens = append(tokens, migrationExprToken{typ: migrationExprTokenAnd, val: value})
				case "or":
					tokens = append(tokens, migrationExprToken{typ: migrationExprTokenOr, val: value})
				case "not":
					tokens = append(tokens, migrationExprToken{typ: migrationExprTokenNot, val: value})
				case "eq", "ne", "contains", "matches", "in":
					tokens = append(tokens, migrationExprToken{typ: migrationExprTokenOp, val: strings.ToLower(value)})
				default:
					if _, err := strconv.ParseFloat(value, 64); err == nil {
						tokens = append(tokens, migrationExprToken{typ: migrationExprTokenNumber, val: value})
					} else {
						tokens = append(tokens, migrationExprToken{typ: migrationExprTokenIdent, val: value})
					}
				}
				continue
			}
			return nil, fmt.Errorf("表达式包含非法字符：%s", input[i:i+1])
		}
	}
	tokens = append(tokens, migrationExprToken{typ: migrationExprTokenEOF})
	return tokens, nil
}

func readMigrationExprString(input string, start int) (string, int, error) {
	quote := input[start]
	var builder strings.Builder
	for i := start + 1; i < len(input); i++ {
		ch := input[i]
		if ch == quote {
			return builder.String(), i + 1, nil
		}
		if ch == '\\' && i+1 < len(input) {
			i++
			switch input[i] {
			case 'n':
				builder.WriteByte('\n')
			case 'r':
				builder.WriteByte('\r')
			case 't':
				builder.WriteByte('\t')
			default:
				builder.WriteByte('\\')
				builder.WriteByte(input[i])
			}
			continue
		}
		builder.WriteByte(ch)
	}
	return "", start, fmt.Errorf("字符串未闭合")
}

func isMigrationExprIdentStart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.'
}

func isMigrationExprIdentPart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.' || r == '-'
}

func (p *migrationExprParser) peek() migrationExprToken {
	if p.pos >= len(p.tokens) {
		return migrationExprToken{typ: migrationExprTokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *migrationExprParser) consume() migrationExprToken {
	token := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return token
}

func (p *migrationExprParser) parseOr() (migrationExprNode, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.peek().typ == migrationExprTokenOr {
		p.consume()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = migrationExprBinaryNode{op: "or", left: left, right: right}
	}
	return left, nil
}

func (p *migrationExprParser) parseAnd() (migrationExprNode, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.peek().typ == migrationExprTokenAnd {
		p.consume()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = migrationExprBinaryNode{op: "and", left: left, right: right}
	}
	return left, nil
}

func (p *migrationExprParser) parseUnary() (migrationExprNode, error) {
	if p.peek().typ == migrationExprTokenNot {
		p.consume()
		child, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return migrationExprNotNode{child: child}, nil
	}
	return p.parsePrimary()
}

func (p *migrationExprParser) parsePrimary() (migrationExprNode, error) {
	if p.peek().typ == migrationExprTokenLParen {
		p.consume()
		node, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if p.peek().typ != migrationExprTokenRParen {
			return nil, fmt.Errorf("缺少右括号")
		}
		p.consume()
		return node, nil
	}
	return p.parseComparison()
}

func (p *migrationExprParser) parseComparison() (migrationExprNode, error) {
	fieldToken := p.consume()
	if fieldToken.typ != migrationExprTokenIdent {
		return nil, fmt.Errorf("期望字段名，实际为：%s", fieldToken.val)
	}
	opToken := p.consume()
	if opToken.typ != migrationExprTokenOp {
		return nil, fmt.Errorf("字段 %s 后缺少运算符", fieldToken.val)
	}
	op := strings.ToLower(opToken.val)
	values, err := p.parseValues(op == "in")
	if err != nil {
		return nil, err
	}
	if err := validateMigrationExpressionField(fieldToken.val, op); err != nil {
		return nil, err
	}
	return migrationExprCompareNode{field: strings.ToLower(fieldToken.val), op: op, values: values}, nil
}

func (p *migrationExprParser) parseValues(expectSet bool) ([]string, error) {
	if expectSet {
		if p.peek().typ != migrationExprTokenLBrace {
			value, err := p.parseSingleValue()
			if err != nil {
				return nil, err
			}
			return []string{value}, nil
		}
		p.consume()
		values := make([]string, 0)
		for p.peek().typ != migrationExprTokenRBrace {
			if p.peek().typ == migrationExprTokenEOF {
				return nil, fmt.Errorf("集合缺少右花括号")
			}
			if p.peek().typ == migrationExprTokenComma {
				p.consume()
				continue
			}
			value, err := p.parseSingleValue()
			if err != nil {
				return nil, err
			}
			values = append(values, value)
			if p.peek().typ == migrationExprTokenComma {
				p.consume()
			}
		}
		p.consume()
		if len(values) == 0 {
			return nil, fmt.Errorf("in 集合不能为空")
		}
		return values, nil
	}
	value, err := p.parseSingleValue()
	if err != nil {
		return nil, err
	}
	return []string{value}, nil
}

func (p *migrationExprParser) parseSingleValue() (string, error) {
	token := p.consume()
	switch token.typ {
	case migrationExprTokenString, migrationExprTokenIdent, migrationExprTokenNumber:
		return token.val, nil
	default:
		return "", fmt.Errorf("期望字段值，实际为：%s", token.val)
	}
}

func validateMigrationExpressionField(field string, op string) error {
	field = normalizeMigrationExpressionField(field)
	allowedOps := map[string]map[string]bool{
		"username": {"==": true, "!=": true, "eq": true, "ne": true, "contains": true, "matches": true, "in": true},
		"email":    {"==": true, "!=": true, "eq": true, "ne": true, "contains": true, "matches": true, "in": true},
		"cah":      {"==": true, "!=": true, "eq": true, "ne": true, "matches": true, "in": true},
		"provider": {"==": true, "!=": true, "eq": true, "ne": true, "matches": true, "in": true},
		"status":   {"==": true, "!=": true, "eq": true, "ne": true, "in": true},
		"role":     {"==": true, "!=": true, "eq": true, "ne": true, "in": true},
		"group":    {"==": true, "!=": true, "eq": true, "ne": true, "contains": true, "matches": true, "in": true},
	}
	ops, ok := allowedOps[field]
	if !ok {
		return fmt.Errorf("不支持的迁移筛选字段：%s", field)
	}
	if !ops[op] {
		return fmt.Errorf("字段 %s 不支持运算符 %s", field, op)
	}
	return nil
}

func (n migrationExprBinaryNode) eval(user model.User) (bool, error) {
	left, err := n.left.eval(user)
	if err != nil {
		return false, err
	}
	if n.op == "and" {
		if !left {
			return false, nil
		}
		return n.right.eval(user)
	}
	if left {
		return true, nil
	}
	return n.right.eval(user)
}

func (n migrationExprNotNode) eval(user model.User) (bool, error) {
	value, err := n.child.eval(user)
	if err != nil {
		return false, err
	}
	return !value, nil
}

func (n migrationExprCompareNode) eval(user model.User) (bool, error) {
	fieldValues := migrationExpressionFieldValues(user, n.field)
	if len(fieldValues) == 0 {
		fieldValues = []string{""}
	}

	matchAny := func(op string, value string) (bool, error) {
		for _, fieldValue := range fieldValues {
			matched, err := compareMigrationExpressionValue(fieldValue, op, value)
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		}
		return false, nil
	}

	switch n.op {
	case "!=", "ne":
		for _, value := range n.values {
			matched, err := matchAny("==", value)
			if err != nil {
				return false, err
			}
			if matched {
				return false, nil
			}
		}
		return true, nil
	case "in":
		for _, value := range n.values {
			matched, err := matchAny("in", value)
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		}
		return false, nil
	default:
		return matchAny(n.op, n.values[0])
	}
}

func compareMigrationExpressionValue(fieldValue string, op string, expected string) (bool, error) {
	switch op {
	case "==", "eq", "in":
		return strings.EqualFold(strings.TrimSpace(fieldValue), strings.TrimSpace(expected)), nil
	case "!=", "ne":
		return !strings.EqualFold(strings.TrimSpace(fieldValue), strings.TrimSpace(expected)), nil
	case "contains":
		return strings.Contains(strings.ToLower(fieldValue), strings.ToLower(expected)), nil
	case "matches":
		return regexp.MatchString(expected, fieldValue)
	default:
		return false, fmt.Errorf("不支持的运算符：%s", op)
	}
}

func migrationExpressionFieldValues(user model.User, field string) []string {
	switch normalizeMigrationExpressionField(field) {
	case "username":
		return []string{user.Username}
	case "email":
		return []string{user.Email}
	case "cah":
		return []string{user.CAHID}
	case "provider":
		return migrationExpressionProviders(user)
	case "status":
		return migrationExpressionStatusValues(user.Status)
	case "role":
		return migrationExpressionRoleValues(user.Role)
	case "group":
		return []string{user.Group}
	default:
		return nil
	}
}

func normalizeMigrationExpressionField(field string) string {
	field = strings.ToLower(strings.TrimSpace(field))
	if field == "cah_id" {
		return "cah"
	}
	return field
}

func migrationExpressionProviders(user model.User) []string {
	providers := make([]string, 0, 7)
	if strings.TrimSpace(user.GitHubId) != "" {
		providers = append(providers, "github")
	}
	if strings.TrimSpace(user.DiscordId) != "" {
		providers = append(providers, "discord")
	}
	if strings.TrimSpace(user.OidcId) != "" {
		providers = append(providers, "oidc")
	}
	if strings.TrimSpace(user.WeChatId) != "" {
		providers = append(providers, "wechat")
	}
	if strings.TrimSpace(user.TelegramId) != "" {
		providers = append(providers, "telegram")
	}
	if strings.TrimSpace(user.LinuxDOId) != "" {
		providers = append(providers, "linuxdo")
	}
	if strings.TrimSpace(user.Password) != "" {
		providers = append(providers, "password")
	}
	return providers
}

func migrationExpressionStatusValues(status int) []string {
	values := []string{strconv.Itoa(status)}
	switch status {
	case common.UserStatusEnabled:
		values = append(values, "enabled")
	case common.UserStatusDisabled:
		values = append(values, "disabled")
	}
	return values
}

func migrationExpressionRoleValues(role int) []string {
	values := []string{strconv.Itoa(role)}
	switch role {
	case common.RoleGuestUser:
		values = append(values, "guest")
	case common.RoleCommonUser:
		values = append(values, "common")
	case common.RoleSupportUser:
		values = append(values, "support")
	case common.RoleAdminUser:
		values = append(values, "admin")
	case common.RoleRootUser:
		values = append(values, "root")
	}
	return values
}
