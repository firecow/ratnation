import {Logger} from "../logger.mjs";
import {State} from "../state-handler.mjs";
import {Ticker} from "../ticker.mjs";

export class CouncilStateCleaner extends Ticker {

    private readonly state;

    constructor ({state}: {state: State; logger: Logger}) {
        super({interval: 1000, tick: async () => this.runCleaner()});
        this.state = state;
    }

    async runCleaner () {
        for (const [i, k] of this.state.kings.entries()) {
            if (k.beat > Date.now() - 10000) continue;
            this.state.kings.splice(i, 1);
            console.log("old king", i);
        }

        for (const [i, k] of this.state.lings.entries()) {
            if (k.beat > Date.now() - 10000) continue;
            this.state.lings.splice(i, 1);
            console.log("old ling", i);
        }

        for (const [i, s] of this.state.services.entries()) {
            const king = this.state.kings.find(k => k.host === s.host && k.bind_port === s.bind_port);
            const ling = this.state.lings.find(l => l.ling_id === s.ling_id);
            if (king && ling) continue;
            this.state.services.splice(i, 1);
        }

        return Promise.resolve();
    }
}
