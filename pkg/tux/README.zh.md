# tux 设计说明

tux 是一个用 Go 编写终端应用界面的框架，全称是 Terminal User Experience。

tux 的基本想法很简单：**界面结构写在 XML 里，状态和行为写在 Go 里，XML 编译器把界面结构生成普通 Go 代码。**

这份文档面向第一次接触 tux 的人和 agent。阅读顺序是：先看工作流，再看 XML 如何编译，最后看框架内部原理。

## 1. 工作流

使用 tux 时，一个组件通常由两个文件组成：

- 一个 Go 文件：保存状态、构造函数和事件处理方法。
- 一个 XML 文件：描述这个组件的界面结构。

例如我们要写一个页面，它显示标题，并有一个输入框可以修改名字。

Go 文件负责状态和行为：

```go
package components

import "github.com/open-portfolios/codefolio/pkg/tux"

type home struct {
    tux.Composite

    title *tux.State[string]
    name  *tux.State[string]
}

type HomeProps struct{}

func Home(props HomeProps) tux.Component {
    return &home{
        title: tux.NewState("Welcome"),
        name:  tux.NewState(""),
    }
}

```

XML 文件负责界面结构：

```xml
<component package="components" name="home">
  <Box str="page">
    <Text str="@{self.title}" />
    <Input state="{self.name}" />
  </Box>
</component>
```

XML 编译器生成 `*_tux.go`：

```go
package components

import (
    tux "github.com/open-portfolios/codefolio/pkg/tux"
    builtin "github.com/open-portfolios/codefolio/pkg/tux/builtin"
)

func (self *home) Build(ctx tux.BuildContext) tux.Component {
    return builtin.Container(
        builtin.ContainerProps{},
        builtin.Box(
            builtin.BoxProps{
                Str: "page",
            },
            builtin.Text(
                builtin.TextProps{
                    Str: self.title.Get(ctx),
                },
            ),
            builtin.Input(
                builtin.InputProps{
                    State: self.name,
                },
            ),
        ),
    )
}
```

用户不需要手写 `Build(ctx)`。它由 XML 生成。

即使组件没有 XML props，也仍然使用空的 props struct，例如 `HomeProps{}`。这样所有组件构造函数都有明确的 props 参数。

## 2. 为什么要分成 XML 和 Go

Go 很适合写状态、事件和业务逻辑，但 Go 没有命名参数。如果直接用 Go 函数调用描述界面，参数顺序会很难维护。

tux 使用 `XxxProps` 结构体作为参数中间层：

```go
builtin.Input(builtin.InputProps{
    State: self.name,
})
```

XML 属性天然是命名的：

```xml
<Input state="{self.name}" />
```

因此 tux 的 XML 编译器只需要把 XML 属性转换成 `XxxProps` 字段。这样 XML 可以保持可读，生成后的 Go 代码也能被 Go 编译器检查。

## 3. 文件约定

一个 XML 文件只允许一个 `<component>` 根节点，并生成一个 `*_tux.go` 文件。

例如：

```text
home.go
home.tux.xml
home_tux.go
```

`home.go` 是用户手写文件。

`home.tux.xml` 是用户手写界面文件。

`home_tux.go` 是编译器生成文件。

## 4. 包路径

tux 根包固定为：

```go
github.com/open-portfolios/codefolio/pkg/tux
```

内置组件放在：

```go
github.com/open-portfolios/codefolio/pkg/tux/builtin
```

XML 编译器会自动注入 tux 根包 import 和 builtin import：

```go
import tux "github.com/open-portfolios/codefolio/pkg/tux"
import builtin "github.com/open-portfolios/codefolio/pkg/tux/builtin"
```

内置组件不需要在 XML 中显式 `<import>`，可以直接使用：

```xml
<Text str="hello" />
<Input state="{self.name}" />
<Container />
```

XML 中显式 import 的非内置组件包会被命名为 `p1`、`p2`、`p3` 等：

```go
import p1 "myapp/components"
```

这样用户组件包 alias 不会和根包 alias `tux`、内置组件 alias `builtin` 冲突。

## 5. XML 根节点

根节点必须是小写的 `<component>`：

```xml
<component package="components" name="home">
  ...
</component>
```

含义：

- `package="components"` 表示生成文件使用 `package components`。
- `name="home"` 表示给 `*home` 生成 `Build(ctx)` 方法。

生成结果：

```go
package components

func (self *home) Build(ctx tux.BuildContext) tux.Component {
    ...
}
```

`<component>` 是 tux 元节点，不是普通界面组件。

## 6. XML import

XML 使用 `<import>` 声明非内置组件。内置组件来自 `github.com/open-portfolios/codefolio/pkg/tux/builtin`，编译器会自动导入，不需要写 `<import>`。

```xml
<import package="myapp/components" component="Box" alias="MyBox" />
```

规则：

- `package` 是 Go import path。
- `component` 是 Go 构造函数名。
- `alias` 是 XML tag 名。
- 如果没有 `alias`，XML tag 等于 `component`。
- 相同 package path 复用同一个 `pN` alias。

例如用户在自己的包里也定义了一个 `Box` 组件。为了避免和内置 `<Box>` 冲突，XML 中用 `alias="MyBox"` 把它命名为 `<MyBox>`：

```go
package components

import "github.com/open-portfolios/codefolio/pkg/tux"

type BoxProps struct {
    Title string
}

func Box(props BoxProps, children ...tux.Component) tux.Component {
    // handwritten component
}
```

另一个 XML 组件想使用它，可以这样 import：

```xml
<component package="pages" name="dashboard">
  <import package="myapp/components" component="Box" alias="MyBox" />

  <MyBox title="Projects">
    <Text str="hello" />
  </MyBox>
</component>
```

生成代码会给 `myapp/components` 分配一个 `pN` alias：

```go
package pages

import (
    tux "github.com/open-portfolios/codefolio/pkg/tux"
    builtin "github.com/open-portfolios/codefolio/pkg/tux/builtin"
    p1 "myapp/components"
)

func (self *dashboard) Build(ctx tux.BuildContext) tux.Component {
    return builtin.Container(
        builtin.ContainerProps{},
        p1.Box(
            p1.BoxProps{
                Title: "Projects",
            },
            builtin.Text(
                builtin.TextProps{
                    Str: "hello",
                },
            ),
        ),
    )
}
```

这里生成的是 `p1.Box(...)`，不是 `p1.MyBox(...)`。`alias` 只影响 XML tag 名，不影响 Go 构造函数名。

如果没有命名冲突，也可以不写 `alias`：

```xml
<import package="myapp/components" component="Box" />

<Box title="Projects" />
```

上面的 import 会形成这样的映射：

| XML tag | Go call        |
| ------- | -------------- |
| `Box`   | `builtin.Box`  |
| `Text`  | `builtin.Text` |
| `MyBox` | `p1.Box`       |

## 7. 组件节点

除了 tux 元节点，其他 XML 节点都表示组件调用。

组件节点必须大写，因为它对应 Go 中 exported constructor：

```xml
<Box>
  <Text str="hello" />
</Box>
```

会生成：

```go
builtin.Container(
    builtin.ContainerProps{},
    builtin.Box(
        builtin.BoxProps{},
        builtin.Text(
            builtin.TextProps{
                Str: "hello",
            },
        ),
    ),
)
```

小写节点只保留给 tux 元节点。第一版有 `<component>`、`<import>` 和 `<children />`。

## 8. Container

`Container` 是 `tux/builtin` 中的基础内置组件。

它没有样式、布局、边框、颜色、padding 等视觉语义。它只负责把多个组件包装成一个组件，并按顺序持有这些子组件。

XML 编译器会让每个组件的 `Build(ctx)` 默认返回一个 `builtin.Container(...)`。

例如：

```xml
<component package="components" name="example">
  <Text str="hello" />
</component>
```

生成：

```go
func (self *example) Build(ctx tux.BuildContext) tux.Component {
    return builtin.Container(
        builtin.ContainerProps{},
        builtin.Text(
            builtin.TextProps{
                Str: "hello",
            },
        ),
    )
}
```

空组件也会生成一个空 `Container`：

```xml
<component package="components" name="empty">
</component>
```

生成：

```go
func (self *empty) Build(ctx tux.BuildContext) tux.Component {
    return builtin.Container(
        builtin.ContainerProps{},
    )
}
```

因此 `Build(ctx)` 总是返回一个组件，不需要为“没有子组件”做特殊分支。

## 9. children 插槽

`<children />` 用于在组合组件 XML 中指定外部子组件应该插入到哪里。

例如定义一个 `Panel` 组合组件：

```xml
<component package="components" name="panel">
  <Box str="panel">
    <Text str="@{self.title}" />
    <children />
  </Box>
</component>
```

用户手写构造函数时接收 children：

```go
type PanelProps struct {
    Title string
}

type panel struct {
    tux.Composite

    title    *tux.State[string]
    children []tux.Component
}

func Panel(props PanelProps, children ...tux.Component) tux.Component {
    return &panel{
        title:    tux.NewState(props.Title),
        children: children,
    }
}
```

XML 编译器会把 `<children />` 生成成一个 `builtin.Container(...)`，并把 `self.children...` 放进这个容器：

```go
func (self *panel) Build(ctx tux.BuildContext) tux.Component {
    return builtin.Container(
        builtin.ContainerProps{},
        builtin.Box(
            builtin.BoxProps{
                Str: "panel",
            },
            builtin.Text(
                builtin.TextProps{
                    Str: self.title.Get(ctx),
                },
            ),
            builtin.Container(
                builtin.ContainerProps{},
                self.children...,
            ),
        ),
    )
}
```

规则：

- `<children />` 是 tux 元节点，不是普通组件。
- `<children />` 必须是自闭合标签。
- `<children />` 只能出现在组件 children 列表的位置。
- `<children />` 生成 `builtin.Container(builtin.ContainerProps{}, self.children...)`。
- 因为 `<children />` 会变成普通组件，所以它可以出现在 children 列表的任意位置。
- 如果组合组件不需要透传外部 children，就不写 `<children />`。

## 10. Props 编译规则

每个组件构造函数至少接收 props：

```go
func Xxx(props XxxProps) tux.Component
```

如果组件需要接收子组件，再添加 variadic children 参数：

```go
func Xxx(props XxxProps, children ...tux.Component) tux.Component
```

即使组件没有任何 props，也必须有一个空的 `XxxProps`：

```go
type XxxProps struct{}
```

调用时传入 `XxxProps{}`。

构造函数也可以返回具体类型，只要该类型实现 `tux.Component`：

```go
func Xxx(props XxxProps) *xxx
```

XML props 必须是 camelCase。

Go `XxxProps` 字段名由 XML prop 首字符大写得到：

| XML prop    | Go field    |
| ----------- | ----------- |
| `str`       | `Str`       |
| `value`     | `Value`     |
| `onChange`  | `OnChange`  |
| `autoFocus` | `AutoFocus` |
| `maxWidth`  | `MaxWidth`  |

示例：

```xml
<Input state="{self.name}" autoFocus="{true}" />
```

生成：

```go
builtin.Input(
    builtin.InputProps{
        State:     self.name,
        AutoFocus: true,
    },
)
```

编译器不需要知道 `InputProps` 是否存在，也不需要检查字段类型。最终由 Go 编译器做类型检查。

## 11. 属性值编译规则

属性值有三种形式。

第一种是普通字符串：

```xml
<Text str="hello" />
```

生成：

```go
builtin.Text(builtin.TextProps{
    Str: "hello",
})
```

第二种是普通 Go 表达式：

```xml
<Button onPress="{self.submit}" />
```

生成：

```go
builtin.Button(builtin.ButtonProps{
    OnPress: self.submit,
})
```

第三种是 State 读取语法糖：

```xml
<Text str="@{self.title}" />
```

生成：

```go
builtin.Text(builtin.TextProps{
    Str: self.title.Get(ctx),
})
```

总结：

| XML value   | Go output       |
| ----------- | --------------- |
| `"content"` | `"content"`     |
| `"{expr}"`  | `expr`          |
| `"@{expr}"` | `expr.Get(ctx)` |

限制：

- `{...}` 必须占据完整属性值。
- `@{...}` 必须占据完整属性值。
- 不支持字符串插值。
- `@{expr}` 只是机械生成 `expr.Get(ctx)`。
- 编译器不检查 `expr` 是否是 `*tux.State[T]`。
- 复杂计算应该写成 Go 方法，然后用 `{self.method()}` 引用。

## 12. XML 不提供控制流

XML 只描述静态组件结构，不提供 `if`、`for` 或任意控制流。

如果需要动态逻辑，写 Go 代码。

例如不要在 XML 中写条件分支，而是写一个手写组件或 Go 方法，然后让 XML 引用它。

这样做的目的：

- XML 编译器保持简单。
- 编译器不需要理解 Go 类型。
- 动态逻辑留在 Go 中，调试更直接。
- 生成代码更可预测。

## 13. Props 只是传参中间层

`Props` 不是组件状态，只是构造函数的参数容器。

组件 struct 不能直接保存 props：

```go
// 不允许
type input struct {
    props InputProps
}
```

组件必须把 props 拆成自己的字段：

```go
// 允许
type input struct {
    state *tux.State[string]
}

func Input(props InputProps) tux.Component {
    return &input{
        state: props.State,
    }
}
```

原因是 props 只是调用构造函数时使用的传参结构。组件实例应该维护自己的明确字段，而不是长期持有 props 对象。

## 14. State 和事件

`State[T]` 保存响应式状态：

```go
type State[T any] struct {
    // internal
}

func NewState[T any](initial T) *State[T]
func (s *State[T]) Get(ctx BuildContext) T
func (s *State[T]) Set(v T)
func (s *State[T]) Update(fn func(T) T)
```

`Get(ctx)` 用来读取状态，并建立订阅关系。

`Set(v)` 和 `Update(fn)` 用来修改状态，并触发订阅者重建。

多数展示组件通过 `@{self.xxx}` 读取 state，也就是生成 `self.xxx.Get(ctx)`。`Input` 这类常用双向绑定组件可以直接接收 `*tux.State[string]`，所以使用普通表达式 `{self.name}` 传递 state 对象。输入内容变化时，`Input` 会直接更新这个 state。

例如：

```xml
<Input state="{self.name}" />
```

等价于：

```go
builtin.Input(builtin.InputProps{
    State: self.name,
})
```

数据流是：

```text
Build(ctx) 读取 State，或把 State 传给支持绑定的组件
  -> 原子组件获得 props
  -> 用户输入触发 Input 内部更新 State
  -> State dirty 它的 subscribers
  -> 相关组件重新 Build
```

`Render(...)` 阶段会接收 `RenderContext` 并向其中绘制内容。它不应该建立 state 订阅；会影响组件结构或 props 的 state 读取应该发生在 `Build(ctx)` 中。

## 15. Component 原理

tux 中所有组件都实现同一个接口：

```go
type Component interface {
    Build(ctx BuildContext) Component
    Render(build BuildContext, render RenderContext) error
}
```

这个接口表示两类组件。

第一类是组合组件。组合组件不直接渲染，它通过 `Build(ctx)` 返回子组件树。

第二类是原子组件。原子组件通过 `Render(...)` 直接向 `RenderContext` 绘制。

`RenderContext` 是底层渲染目标：

```go
type RenderContext interface {
    Paint(row, column int, b byte)
    Flush() error
}
```

当前 debug renderer 用二维 byte buffer 实现这个接口。未来真实 terminal renderer 可以用同一个接口实现 terminal cell、样式、diff 和 flush。

运行时规则是：

- 组合组件通过 `Build(ctx)` 返回另一个组件，通常继承 `Composite.Render(...)` 的 no-op 行为。
- 原子组件的 `Build(ctx)` 返回 nil，并实现 `Render(...)` 向渲染上下文绘制。
- 拥有 children 的组件，例如 `Container`，自己决定如何展开和渲染这些 children。
- 如果 `Build(ctx)` 返回组件自身，tux 认为这是 composition cycle，并直接 panic。

## 16. Atomic 与 Composite

tux 提供两个辅助嵌入类型。

原子组件嵌入 `tux.Atomic`：

```go
type Atomic struct{}

func (Atomic) Build(ctx BuildContext) Component {
    return nil
}

func (Atomic) Render(build BuildContext, render RenderContext) error {
    panic("tux: atomic component must implement Render")
}
```

组合组件嵌入 `tux.Composite`：

```go
type Composite struct{}

func (Composite) Build(ctx BuildContext) Component {
    panic("tux: composite component must implement Build")
}

func (Composite) Render(build BuildContext, render RenderContext) error {
    return nil
}
```

这样可以保证：

- 原子组件如果忘记实现 `Render(...)`，运行时会 panic。
- 组合组件如果没有生成 `Build(ctx)`，运行时会 panic。
- 原子组件默认没有子组件展开逻辑。
- 组合组件默认没有直接渲染逻辑。

## 17. 原子组件

原子组件是手写 Go 组件，负责直接向 `RenderContext` 绘制。

示例：

```go
type TextProps struct {
    Row    int
    Column int
    Str    string
}

type text struct {
    tux.Atomic

    row    int
    column int
    str    string
}

func Text(props TextProps) tux.Component {
    return &text{
        row:    props.Row,
        column: props.Column,
        str:    props.Str,
    }
}

func (t *text) Render(build tux.BuildContext, render tux.RenderContext) error {
    for i := 0; i < len(t.str); i++ {
        render.Paint(t.row, t.column+i, t.str[i])
    }
    return nil
}
```

原子组件可以接收 children。是否使用 children 由该组件自己的 `Render(...)` 决定。

`Container` 是这个规则的第一个内置例子：它是 atomic，但它的 render 方法会对每个 child 调用 `Build(ctx)`，直到展开到 atomic artifact，然后调用这个 child 的 `Render(...)`。如果某个 child build 返回自身，`Container` 会 panic，因为这是 composition cycle。

## 18. 组合组件

组合组件通常由用户手写状态 struct，XML 生成 `Build(ctx)`。

手写部分：

```go
type home struct {
    tux.Composite

    title *tux.State[string]
}
```

生成部分：

```go
func (self *home) Build(ctx tux.BuildContext) tux.Component {
    return builtin.Container(
        builtin.ContainerProps{},
        builtin.Text(builtin.TextProps{
            Str: self.title.Get(ctx),
        }),
    )
}
```

`Build(ctx)` 可以多次调用。它是“当前状态到组件树”的投影函数。

## 19. 局部重建

tux 不使用 whole-app dirty，也就是不会因为一个状态变化就把整个应用都标记为需要重建。

这里的 dirty 表示“这个 mounted element 依赖的数据变了，需要重新执行 `Build(ctx)`”。

`State.Get(ctx)` 会订阅当前 `BuildContext` 对应的 mounted element。

`State.Set(...)` 只 dirty 这些 subscribers。

状态更新链路：

```text
State.Set(...)
  -> dirty subscribers
  -> scheduler rebuild dirty elements
  -> Build(ctx) 重新生成组件树
  -> reconcile child component
  -> atomic Render(...) 绘制到 RenderContext
  -> buffer diff / terminal flush
```

subscriber 绑定的是运行时内部的 mounted element，不是临时创建出来的 component value。因为 `Build(ctx)` 每次都可以创建新的 component tree，component value 本身不适合作为稳定身份。

## 20. Key

tux 可以拥有 key 机制，但 key 由框架维护，用户不需要手动指定。

key 用于 mounted element 的 identity 和 reconcile。用户 XML 中不需要写 `key` 作为常规 props。

后续如果需要列表或动态 children 的稳定身份，可以单独设计显式 key 机制。

## 21. 不做的事情

第一版明确不做：

- 不做 GSX。
- 不在 XML 中写 Go 函数体。
- 不在 XML 中提供 `if` / `for` 控制流。
- 不做字符串插值。
- 不要求 XML 编译器读取外部 Go 类型。
- 不要求 XML 编译器检查 props 字段是否存在。
- 不要求 XML 编译器检查 `@{...}` 是否是 state。
- 不使用 whole-app dirty。
- 不允许 `Render(...)` 阶段参与 state 订阅。
