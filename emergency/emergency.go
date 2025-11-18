package emergency

/*
Emergency:
```yaml
type: emergency
payload:
    value: true | false

type: mediate
payload:
    topic: Remote/#
    value: true | false
```

Emergency copies what happens on Remote onto its remote branch.
When `emergency=true`, then it stops mirroring.
It will send a speed and velocity signal, which the cars are subscribed to.
*/
