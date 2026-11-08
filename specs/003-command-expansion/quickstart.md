# Quickstart / Validation: Command expansion

Sandbox `HOME`, block installed (`ali init zsh`, upgraded via feature 002). Interactive checks run through the `zpty` driver (research R10).

```sh
ali add k kubectl
ali sub add kubectl g get; ali sub add kubectl p pods; ali sub add kubectl n '-n {{ns}}'  # n fails until ns exists
ali var add ns one=namespace-1 sys=kube-system --default one
ali sub add kubectl n '-n {{ns}}'
ali sub add --shared o '-o wide'; ali sub add helm g 'get values'
ali env add dev ns=one; ali env add prod ns=prod --protect
```

| Type (then Enter) | Expected command line / result |
|---|---|
| `k g p ` | line shows `kubectl get pods ` before Enter |
| `k g p n @sys` | runs `kubectl get pods -n kube-system`; history has that text |
| `k g p n` | `-n namespace-1` (default) |
| `ali env use prod` then `k g p n` | first Enter shows `kubectl get pods -n prod` + protected notice; second Enter runs; prompt shows `prod` in red |
| second terminal | still default; `ali env use --global dev` → both follow `dev` except a terminal with its own choice |
| `k get p` / `k logs mypod p` | `p` expands to `pods` |
| `k logs \p` | stays `kubectl logs p` |
| `k exec mypod -- g` | `g` not expanded after `--` |
| `ali help expansion`, `man ali-sub` | documented with examples |
| `npm i @types/node` | untouched, no process spawned (check with a counter in the test) |
| `k g p n @ns='a b; rm -rf ~'` | argument reaches kubectl as one literal word |
| `ali config set expand runtime`; `k g p n @sys` | line unchanged while typing; same command runs |
| `ali expand 'k g p o'` | prints `kubectl get pods -o wide` |
| `ali show kubectl` | lists k, g, p, n, shared o, variable ns |

Also: SC-003 timing (typing loop with/without trigger words), and a manual check with zsh-autosuggestions and zsh-syntax-highlighting installed.
