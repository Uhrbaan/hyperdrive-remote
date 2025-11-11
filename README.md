# 
> **Steering**: One gives a directive $D$ which directly has an effect on the result. 
>
> **Controlling**: Onve gives a directive $D$, which is more of a _goal_. The system then reads the current stat of the system and adapts its own state to achieve the directive.

## Emergency 
```yaml
type: emergency
payload:
    value: true | false

type: mediate 
payload: 
    topic: Remote/#
    value: true | false
```

Emergency is a _conditional mediator_: it relays the values of the remote controller identically, unless the emergency is sest to `true`.