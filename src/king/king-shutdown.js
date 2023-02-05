import wait from "wait-promise";

export function initKingShutdownHandlers({context, stateHandler, syncer, ratholeManager}) {
    const listener = async(signal) => {
        console.log("msg=\"king shutdown sequence initiated\" service.type=ratking");
        context.shuttingDown = true;
        stateHandler.stop();
        await syncer.stop();
        await wait.sleep(1000);
        ratholeManager.killProcesses(signal);
    };
    process.on("SIGINT", listener);
    process.on("SIGTERM", listener);
}
