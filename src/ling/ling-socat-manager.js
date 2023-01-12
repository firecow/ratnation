import {execa} from "execa";

export class LingSocatManager {
    socapProcessMap = new Map();

    constructor(context) {
        this.context = context;
    }

    #each(socatCnf) {
        const state = this.context.state;
        const config = this.context.config;
        const bindPort = socatCnf["bind_port"];
        const name = socatCnf["name"];
        if (this.socapProcessMap.has(bindPort)) return;

        const services = state["services"].filter(s => config.socatMap.has(s["name"]) && s["king_ready"] && s["ling_ready"]);
        const service = services.find(s => s["name"] === name);
        if (!service) {
            return console.warn(`msg="Could not find '${name}' service in council state" service_type=ratling log.logger=socat-manager`);
        }

        const socat = execa("socat", [`tcp-l:${bindPort},fork,reuseaddr`, `tcp:${service["host"]}:${service["remote_port"]}`], {cwd: "src/ling/"});
        console.log(`msg="Started socat" service_type=ratling log.logger=socat-manager port=${bindPort} pid=${socat.pid} host=${service["host"]} remote_port=${service["remote_port"]}`);
        socat.stdout.pipe(process.stdout);
        socat.stderr.pipe(process.stderr);
        socat.on("exit", (code) => {
            console.info(`msg="Socat exited" process_exit_code=${code} service_type=ratling log.logger=socat-manager`);
            this.socapProcessMap.delete(bindPort);
            this.doit();
        });
        socat.on("error", (err) => {
            console.error(err.message);
            process.exit(1);
        });
        this.socapProcessMap.set(bindPort, socat);
    }

    async doit() {
        for (const socatCnf of this.context.config.socatMap.values()) {
            this.#each(socatCnf);
        }
    }
}
