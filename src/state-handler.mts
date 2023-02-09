import got from "got";
import {to} from "await-to-js";
import {Ticker} from "./ticker.mjs";

export interface StateKing {
    bind_port: number;
    host: string;
    ports: string;
    shutting_down: boolean;
    beat: number;
    location: string;
}

export interface StateLing {
    ling_id: string;
    shutting_down: boolean;
    beat: number;
}

export interface StateService {
    name: string;
    token: string;
    service_id: string;
    ling_id: string;
    prefered_location: string;

    ling_ready: boolean;
    king_ready: boolean;

    host: string | null;
    bind_port: number | null;
    remote_port: number | null;
}

export interface State {
    services: StateService[];
    kings: StateKing[];
    lings: StateLing[];
    revision: number;
}

interface StateHandlerOpts {
    councilHost: string;
    updatedFunc: (state: State) => Promise<void> | void;
}

export class StateHandler extends Ticker {

    private readonly updatedFunc;
    private readonly councilHost;
    private state: State | null = null;

    constructor ({councilHost, updatedFunc}: StateHandlerOpts) {
        super({interval: 500, tick: async () => await this.fetchState()});
        this.updatedFunc = updatedFunc;
        this.councilHost = councilHost;
    }

    hasState () {
        return this.state !== null;
    }

    async fetchState () {
        const [err, response] = await to(got.get(`${this.councilHost}/state`));
        if (err || response.statusCode !== 200) {
            return console.error("Failed to fetch state from council", err?.message ?? response?.statusCode ?? 0);
        }

        const newState = JSON.parse(response.body) as State;
        if (this.state === null || this.state.revision !== newState.revision) {
            this.state = newState;
            await this.updatedFunc(newState);
        }
    }
}
