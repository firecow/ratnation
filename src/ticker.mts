

export class Ticker {

    private readonly interval;
    private readonly tick;
    private timer?: NodeJS.Timeout;

    constructor ({interval, tick}: {interval: number; tick: () => Promise<void>}) {
        this.interval = interval;
        this.tick = tick;
    }

    start () {
        const now = Date.now();
        void this.tick().then(() => {
            const delta = Date.now() - now;
            this.timer = setTimeout(() => this.start(), Math.max(this.interval - delta, 0));
        });
    }

    async stop () {
        clearTimeout(this.timer);
        await this.tick();
    }
}
