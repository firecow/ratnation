export class Provisioner {

    constructor({state}) {
        this.state = state;
    }

    #getUnusedPort(state, king) {
        const from = Number(king.ports.split("-")[0]);
        const to = Number(king.ports.split("-")[1]);
        const used = state["services"].filter(s => s["bind_port"] === king["bind_port"] && s["host"] === king["host"]).map(s => s["remote_port"]);
        const unused = [];
        for (let i = from; i <= to; i++) {
            if (used.includes(i)) continue;
            unused.push(i);
        }

        return unused.random();
    }

    #eachUnprovisioned(state, service) {
        let port, king, retries = 100;
        do {
            king = state.kings.filter(k => k.shutting_down === false).random();
            if (!king) return console.warn(`msg="Could not find suited king for ${service.name}" service.type=ratcouncil`);
            port = this.#getUnusedPort(state, king);
            retries--;
        } while (port == null && retries !== 0);

        if (port == null) {
            return console.error(`msg="Did not find available remote_port on any kings for ${service.name}" service.type=ratcouncil`);
        }

        service.location = king.location;
        service.host = king.host;
        service.remote_port = port;
        service.bind_port = king.bind_port;

        state.revision++;

        console.log(`msg="'${service.name}' provisioned to ${king.host}:${service.bind_port}, exposed on ${king.host}:${service.remote_port}" service.type=ratcouncil`);
    }

    async provision() {
        const unprovisioned = this.state.services.filter(s => s.bind_port === null);
        unprovisioned.forEach(u => this.#eachUnprovisioned(this.state, u));
    }
}
