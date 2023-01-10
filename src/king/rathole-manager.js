import fs from "fs";
import {execa} from "execa";
import got from "got";

export class RatholeManager {
    ratholeProcessMap = new Map();

    async doit({host, config, state, councilHost}) {
        for (const ratholeCnf of config.ratholeMap.values()) {
            const bindPort = ratholeCnf.bind_port;
            const services = state.services.filter(s => s["bind_port"] === ratholeCnf.bind_port);
            if (services.length === 0) continue;

            const lines = [
                "[server]",
                `bind_addr = "0.0.0.0:${bindPort}"`,
                "",
                "[server.services]",
            ];

            for (const service of services) {
                lines.push(`[server.services.${service["name"]}]`);
                lines.push(`token = "${service["token"]}"`);
                lines.push(`bind_addr = "0.0.0.0:${service["remote_port"]}"`);
            }

            const ratholeConfig = `server-${bindPort}.toml`;
            await fs.promises.writeFile(`src/king/${ratholeConfig}`, `${lines.join("\n")}\n`, "utf8");
            await got(`${councilHost}/king-active`, {
                method: "PUT",
                json: {
                    names: services.map(s => s["name"]),
                },
            });

            if (this.ratholeProcessMap.has(bindPort)) continue;

            const rathole = execa("rathole", [ratholeConfig], {cwd: "src/king/"});
            console.log(`Started rathole bind_addr=${bindPort} pid=${rathole.pid}`);
            rathole.stdout.pipe(process.stdout);
            rathole.stderr.pipe(process.stderr);
            rathole.on("exit", () => {
                console.info("Closed rathole");
                this.ratholeProcessMap.delete(bindPort);
                this.doit({host, councilHost, config, state});
            });

            this.ratholeProcessMap.set(bindPort, rathole);
        }
    }
}
