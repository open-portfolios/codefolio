package components

import tui "github.com/grindlemire/go-tui"

templ Composer(r *tui.Ref) {
    <div ref={r} class="p-1 bg-[#212121]">
        {children...}
    </div>
}
