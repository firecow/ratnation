import {Logger} from "../logger.js";
import {State} from "../state-handler.js";
import {Ticker} from "../ticker.js";
import {Server as SocketIoServer} from "socket.io";

export class CouncilStateCleaner extends Ticker {

    private readonly state;
    private readonly socketIo: SocketIoServer;

    constructor ({state, socketIo}: {state: State; socketIo: SocketIoServer; logger: Logger}) {
        super({interval: 1000, tick: async () => this.runCleaner()});
        this.state = state;
        this.socketIo = socketIo;
    }

    async runCleaner () {
        let stateChanged = false;

        for (const [i, k] of this.state.kings.entries()) {
            if (k.beat > Date.now() - 10000) continue;
            this.state.kings.splice(i, 1);
            stateChanged = true;
        }

        for (const [i, k] of this.state.lings.entries()) {
            if (k.beat > Date.now() - 10000) continue;
            this.state.lings.splice(i, 1);
            stateChanged = true;
        }

        for (const [i, s] of this.state.services.entries()) {
            const king = this.state.kings.find(k => k.host === s.host && k.bind_port === s.bind_port);
            const ling = this.state.lings.find(l => l.ling_id === s.ling_id);
            if (king && ling) continue;
            this.state.services.splice(i, 1);
            stateChanged = true;
        }

        if (stateChanged) {
            this.state.revision++;
            this.socketIo.sockets.emit("state-changed");
        }

        return Promise.resolve();
    }
}
