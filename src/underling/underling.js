import {Config} from "./config.js";
import {ConfigSender} from "./config-sender.js";
import {StateHandler} from "../state-handler.js";
import {SocatManager} from "./socat-manager.js";
import {RatholeManager} from "./rathole-manager.js";

export const command = "underling";
export const description = "Start ratunderling";

export async function handler(argv) {
    const councilHost = argv["council-host"];
    const config = new Config(argv);
    const configSyncer = new ConfigSender({councilHost, config});
    const socatManager = new SocatManager();
    const ratholeManager = new RatholeManager();
    const stateHandler = new StateHandler({
        councilHost,
        updatedFunc: (state) => {
            socatManager.doit(config, state);
            ratholeManager.doit(config, state);
        },
    });
    stateHandler.start();
    configSyncer.start();

    // socat openssl-listen:4433,for,reuseaddr,cert=$HOME/etc/server.pem,cafile=$HOME/etc/client.crt tcp:127.0.0.1:2020
    // socat tcp:127.0.0.1:2020 openssl-connect:server.domain.org:4433,cert=$HOME/etc/client.pem,cafile=$HOME/etc/server.crt
    // socat tcp-l:5050,fork,reuseaddr tcp:127.0.0.1:2020
}

export function builder(yargs) {
    yargs.options("council-host", {
        type: "string",
        description: "Council host to syncronize from",
        default: "http://localhost:8080",
    });
    yargs.options("rathole", {
        type: "array",
        description: "Rathole clients to open if council state matches",
    });
    yargs.options("socat", {
        type: "array",
        description: "Socats to open based on config if council state matches",
    });
}
