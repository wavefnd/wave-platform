---
translation_set_id: design-goals
path: getting-started/design-goals
locale: en
group: getting-started
group_order: 1
order: 4
title: Language design and guarantees
summary: The design boundaries behind Wave's explicit types, native execution, and evolving implementation.
---

## Design goals

- Make data representation and state changes visible in source.
- Compile native programs without hiding ABI and target decisions.
- Keep low-level facilities composable with readable structures and functions.
- Allow the language and standard library to evolve independently.

## Explicit types

Local variables, fields, parameters, results, arrays, and native boundaries state their types. Wave 0.2.0-pre-beta does not treat omitted type annotations as the normal programming model.

## Implementation status

A reserved token or documented roadmap item is not necessarily an implemented language feature. This manual distinguishes verified release behavior from planned work and reports known release limitations where they matter.

