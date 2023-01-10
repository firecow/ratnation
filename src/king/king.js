import {Config} from "./config.js";
import {StateHandler} from "../state-handler.js";
import {RatholeManager} from "./rathole-manager.js";
import {ConfigSender} from "./config-sender.js";

export const command = "king";
export const description = "Start ratking";

export async function handler(argv) {
    const councilHost = argv["council-host"];
    const host = argv["host"];
    const config = new Config(argv);
    const ratholeManager = new RatholeManager();
    const configSender = new ConfigSender({councilHost, config, host, location: "mylocation"}); // TODO From cli options
    const stateHandler = new StateHandler({
        councilHost,
        updatedFunc: (state) => {
            ratholeManager.doit({host: argv["host"], councilHost, config, state});
        },
    });
    stateHandler.start();
    configSender.start();
}

export function builder(yargs) {
    yargs.options("council-host", {
        type: "string",
        description: "Council host to syncronize from",
        default: "http://localhost:8080",
    });
    yargs.options("host", {
        type: "string",
        description: "Host (domain or ip)",
        demand: true
    });
    yargs.options("rathole", {
        type: "array",
        description: "Rathole servers to open if council state matches",
    });
}
