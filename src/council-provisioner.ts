import assert from "assert";
import {Logger} from "./logger.js";
import {State, StateService} from "./state-handler.js";

export class CouncilProvisioner {

    private readonly logger: Logger;

    constructor ({logger}: {logger: Logger}) {
        this.logger = logger;
    }

    availableKingPorts (state: State) {
        const kings = state.kings.filter(k => !k.shutting_down);

        const kingPorts = [];

        for (const king of kings) {
            const from = Number(king.ports.split("-")[0]);
            const to = Number(king.ports.split("-")[1]);
            const used = state.services.filter(s => s.bind_port === king.bind_port && s.host === king.host).map(s => s.remote_port);
            const ports = [];
            for (let i = from; i <= to; i++) {
                if (used.includes(i)) continue;
                ports.push(i);
            }
            if (ports.length > 0) {
                kingPorts.push({king, ports});
            }
        }

        return kingPorts;
    }

    provisionService (state: State, service: StateService) {
        const logger = this.logger;

        const found = this.availableKingPorts(state).shift();
        if (!found) {
            return logger.error(`No available remote_port found on any kings for ${service.service_id}`, {
                "service.type": "ratcouncil",
            });
        }
        const {ports, king} = found;
        const remotePort = ports.shift();
        assert(remotePort != null, "remote_port cannot be undefined or null");

        service.host = king.host;
        service.remote_port = remotePort;
        service.bind_port = king.bind_port;

        state.revision++;

        logger.info(`Provisioned ${service.name} to ${king.host}:${service.bind_port}, exposed on ${king.host}:${service.remote_port}`, {
            "service.type": "ratcouncil",
        });
    }

    provision (state: State) {
        const unprovisionedServices = state.services.filter(s => s.bind_port === null);
        unprovisionedServices.forEach(unprovisionedService => {
            this.provisionService(state, unprovisionedService);
        });
    }
}
