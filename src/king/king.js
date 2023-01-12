import {KingConfig} from "./king-config.js";
import {StateHandler} from "../state-handler.js";
import {KingRatholeManager} from "./king-rathole-manager.js";
import {KingSyncer} from "./king-syncer.js";
import wait from "wait-promise";

export const command = "king";
export const description = "Start ratking";

export async function handler(argv) {
    const councilHost = argv["council-host"];
    const host = argv["host"];
    const config = new KingConfig(argv);
    const context = {state: null, readyServices: [], config, host, councilHost, location: "mylocation"}; // TODO: location from cli options
    const ratholeManager = new KingRatholeManager(context);
    const kingSyncer = new KingSyncer(context);
    const stateHandler = new StateHandler({
        ...context,
        updatedFunc: (state) => {
            context.state = state;
            ratholeManager.doit();
        },
    });

    stateHandler.start();
    await wait.until(() => stateHandler.hasState());
    kingSyncer.start();
    console.log("msg=\"king ready\" service_type=ratking");
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
