package components

import tui "github.com/grindlemire/go-tui"

templ MainPanel() {
    <div class="h-full grow m-1 flex-col">
        {children...}
    </div>
}
