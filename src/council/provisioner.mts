import {State, StateKing, StateService} from "../state-handler.mjs";
import {Logger} from "../logger.mjs";

export class Provisioner {

    private readonly logger: Logger;
    private readonly state;

    constructor ({state, logger}: {state: State; logger: Logger}) {
        this.logger = logger;
        this.state = state;
    }

    static randomFromArray<T>(arr: T[]): T {
        return arr[Math.floor((Math.random() * arr.length))];
    }

    #getUnusedPort (state: State, king: StateKing) {
        const from = Number(king.ports.split("-")[0]);
        const to = Number(king.ports.split("-")[1]);
        const used = state["services"].filter(s => s["bind_port"] === king["bind_port"] && s["host"] === king["host"]).map(s => s["remote_port"]);
        const unused = [];
        for (let i = from; i <= to; i++) {
            if (used.includes(i)) continue;
            unused.push(i);
        }

        return Provisioner.randomFromArray(unused);
    }

    #provisionService (state: State, service: StateService) {
        const logger = this.logger;
        let port: number, king: StateKing, retries = 100;
        // TODO: Make deterministic host/port provisioning by removing randomess.
        do {
            king = Provisioner.randomFromArray(state.kings.filter(k => !k.shutting_down));
            if (!king) {
                return logger.error(`Could not find suited king for ${service.name}`, {
                    "service.type": "ratcouncil",
                });
            }
            port = this.#getUnusedPort(state, king);
            retries--;
        } while (port == null && retries !== 0);

        if (port == null) {
            return logger.error(`Available remote_port not found on any kings for ${service.name}`, {
                "service.type": "ratcouncil",
            });
        }

        service.host = king.host;
        service.remote_port = port;
        service.bind_port = king.bind_port;

        state.revision++;

        logger.info(`Provisioned ${service.name} to ${king.host}:${service.bind_port}, exposed on ${king.host}:${service.remote_port}`, {
            "service.type": "ratcouncil"
        });
    }

    provision () {
        const unprovisionedServices = this.state.services.filter(s => s.bind_port === null);
        unprovisionedServices.forEach(unprovisionedService => {
            this.#provisionService(this.state, unprovisionedService);
        });
    }
}
