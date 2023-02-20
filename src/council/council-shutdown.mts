import delay from "delay";
import {Server} from "http";
import {Logger} from "../logger.mjs";
import {CouncilStateCleaner} from "./coucil-state-cleaner.mjs";

interface CouncilShutdownHandlersOpts {
    logger: Logger;
    server: Server;
    cleaner: CouncilStateCleaner;
}

let shuttingDown = false;

export function initCouncilShutdownHandlers ({logger, server, cleaner}: CouncilShutdownHandlersOpts) {
    const listener = async () => {
        if (shuttingDown) return;
        shuttingDown = true;
        logger.info("Shutdown sequence initiated", {"service.type": "ratcouncil"});
        await delay(5000);
        cleaner.stop();
        server.close();
    };
    process.on("SIGINT", listener);
    process.on("SIGTERM", listener);
}
