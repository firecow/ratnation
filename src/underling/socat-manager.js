import {execa} from "execa";

export class SocatManager {
    socapProcessMap = new Map();

    async doit(config, state) {

        for (const socatCnf of config.socatMap.values()) {
            const port = socatCnf["port"];
            const name = socatCnf["name"];
            if (this.socapProcessMap.has(port)) continue;

            const services = state["services"].filter(s => config.socatMap.has(s["name"]) && s["king_active"] === true);
            const service = services.find(s => s["name"] === name);
            if (!service) {
                console.warn(`Could not find '${name}' service in council state`);
                continue;
            }

            const socat = execa("socat", [`tcp-l:${port},fork,reuseaddr`, `tcp:${service["host"]}:${service["remote_port"]}`], {cwd: "src/underling/"});
            console.log(`Started socat port=${port} pid=${socat.pid} host=${service["host"]} remote_port=${service["remote_port"]}`);
            socat.stdout.pipe(process.stdout);
            socat.stderr.pipe(process.stderr);
            socat.on("exit", () => {
                console.info("Closed socat");
                this.socapProcessMap.delete(port);
                this.doit(config, state);
            });
            socat.on("error", (err) => {
                console.error(err.message);
                process.exit(1);
            });
            this.socapProcessMap.set(port, socat);
        }
    }
}
