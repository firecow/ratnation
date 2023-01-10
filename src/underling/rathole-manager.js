import {execa} from "execa";
import fs from "fs";

export class RatholeManager {
    ratholeProcessMap = new Map();

    async doit(config, state) {
        const services = state["services"].filter(s => config.ratholeMap.has(s["name"]) && s["king_active"] === true);
        for (const kingBindAddr of services.map(s => `${s["host"]}:${s["bind_port"]}`)) {
            const lines = [
                "[client]",
                `remote_addr = "${kingBindAddr}"`,
                "",
                "[client.services]",
            ];

            for (const service of services.filter(s => `${s["host"]}:${s["bind_port"]}` === kingBindAddr)) {
                const ratholeCnf = Array.from(config.ratholeMap.values()).find(r => r["name"] === service["name"]);
                lines.push(`[client.services.${service["name"]}]`);
                lines.push(`token = "${service["token"]}"`);
                lines.push(`local_addr = "${ratholeCnf["local_addr"]}"`);
            }

            const ratholeConfig = `client-${kingBindAddr.replace(/:/g, "-")}.toml`;
            await fs.promises.writeFile(`src/underling/${ratholeConfig}`, `${lines.join("\n")}\n`, "utf8");

            if (this.ratholeProcessMap.has(kingBindAddr)) continue;

            const rathole = execa("rathole", [ratholeConfig], {cwd: "src/underling/"});
            console.log(`Started rathole remote_addr=${kingBindAddr} pid=${rathole.pid}`);
            rathole.stdout.pipe(process.stdout);
            rathole.stderr.pipe(process.stderr);
            rathole.on("exit", () => {
                console.info("Closed rathole");
                this.ratholeProcessMap.delete(kingBindAddr);
                this.doit(config, state);
            });

            this.ratholeProcessMap.set(kingBindAddr, rathole);
        }
    }
}
