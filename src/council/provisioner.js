
export class Provisioner {

    constructor({state}) {
        this.state = state;
    }

    #getUnusedPort(king) {
        const from = Number(king.ports.split("-")[0]);
        const to = Number(king.ports.split("-")[1]);
        const unused = [];
        for (let i = from; i <= to; i++) {
            if (king.used.includes(i)) continue;
            unused.push(i);
        }

        return unused.random();
    }

    #eachUnprovisioned(state, service) {
        let port, king, retries = 100;
        do {
            king = state.kings.filter(k => k.shutting_down === false).random();
            if (!king) {
                return console.warn(`Could not find suited king for ${service.name}`);
            }

            port = this.#getUnusedPort(king);
            retries--;
        } while(port != null && retries !== 0);

        if (port == null) {
            return console.error(`Did not find available remote_port on any kings for ${service.name}`);
        }

        service.location = king.location;
        service.host = king.host;
        service.remote_port = port;
        service.bind_port = king.bind_port;
        king.used.push(port);

        console.log(`'${service.name}' provisioned to ${king.host}:${service.bind_port}, exposed on ${king.host}:${service.remote_port}`);
    }

    async #provision() {
        const unprovisioned = this.state.services.filter(s => s.bind_port === null);
        unprovisioned.forEach(u => this.#eachUnprovisioned(this.state, u));
    }

    start() {
        this.#provision().then(() => {
            setTimeout(() => this.start(), 100);
        }).catch(err => console.error(err));
    }

}
