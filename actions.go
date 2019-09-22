package onkyoctl

type ISCPGroup string

type Property struct {
    Name  string
    group ISCPGroup
    iscpValue string
}

func (p *Property) Group() ISCPGroup {
    return p.group
}

func (p *Property) Command() ISCPCommand {
    return ISCPCommand(string(p.group) + p.iscpValue)
}

func (p *Property) QueryCommand() ISCPCommand {
    return ISCPCommand(string(p.group) + "QSTN")
}

func (p *Property) Parse(command ISCPCommand) error {
    // match group or error

    // apply value

    return nil
}

// Action ---------------------------------------------------------------------

type Action struct {
    Name string
    group ISCPGroup
    iscpValue string
}

func (a *Action) Command() ISCPCommand {
    return ISCPCommand(string(a.group) + a.iscpValue)
}

// Event ----------------------------------------------------------------------

type Event struct {

}
