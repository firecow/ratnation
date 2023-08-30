import delay from "delay";
import {Server} from "http";
import {Logger} from "../logger.js";
import {CouncilStateCleaner} from "./coucil-state-cleaner.js";

interface CouncilShutdownHandlersOpts {
    logger: Logger;
    httpServer: Server;
    cleaner: CouncilStateCleaner;
}

let shuttingDown = false;

export function initCouncilShutdownHandlers ({logger, httpServer, cleaner}: CouncilShutdownHandlersOpts) {
    const listener = async () => {
        if (shuttingDown) return;
        shuttingDown = true;
        logger.info("Shutdown sequence initiated", {"service.type": "ratcouncil"});
        await delay(5000);
        cleaner.stop();
        httpServer.close();
    };
    process.on("SIGINT", listener);
    process.on("SIGTERM", listener);
}
