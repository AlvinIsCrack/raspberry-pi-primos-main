<script lang="ts">
    interface TickerContent {
        message: string;
        imageSrc?: string;
        badge?: string;
    }

    const currentContent: TickerContent = {
        message: "¡Felices Fiestas Patrias! ^^",
        imageSrc: "/media/chile-flag.png",
    };

    // Parser interno estilo BBCode
    function parseFormattedText(text: string): string {
        return text
            .replace(
                /\[red\](.*?)\[\/red\]/g,
                '<span class="text-danger font-bold">$1</span>',
            )
            .replace(
                /\[blue\](.*?)\[\/blue\]/g,
                '<span class="text-info font-bold">$1</span>',
            )
            .replace(
                /\[white\](.*?)\[\/white\]/g,
                '<span class="text-foreground font-bold">$1</span>',
            )
            .replace(
                /\[warning\](.*?)\[\/warning\]/g,
                '<span class="text-warning font-bold">$1</span>',
            )
            .replace(
                /\[success\](.*?)\[\/success\]/g,
                '<span class="text-success font-bold">$1</span>',
            );
    }

    const renderedMessage = $derived(
        parseFormattedText(currentContent.message),
    );
</script>

<div
    class="absolute bottom-6 left-6 z-20 flex items-center gap-3 rounded border-2 border-border bg-background px-4 py-2 font-mono select-none"
>
    {#if currentContent.imageSrc}
        <img
            src={currentContent.imageSrc}
            alt=""
            class="h-10 w-auto object-contain select-none"
        />
    {:else if currentContent.badge}
        <span class="text-xl">{currentContent.badge}</span>
    {/if}

    <span class="text-base font-medium text-foreground">
        {@html renderedMessage}
    </span>
</div>
