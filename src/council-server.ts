import {AssertionError} from "assert";
import findmyway from "find-my-way";
import http, {IncomingMessage, ServerResponse} from "http";
import {Logger} from "./logger.js";
import {State} from "./state-handler.js";
import {CouncilProvisioner} from "./council-provisioner.js";
import getState from "./routes/council-route-get-state.js";
import putKing from "./routes/council-route-put-king.js";
import putLing from "./routes/council-route-put-ling.js";
import {to} from "./utils.js";
import {Server as SocketIoServer} from "socket.io";

export interface RouteRes { end: (str: string) => void; setHeader: (key: string, val: string) => void}
export interface RouteCtx {
    logger: Logger;
    req: IncomingMessage;
    res: RouteRes;
    provisioner: CouncilProvisioner;
    state: State;
    socketIo: SocketIoServer;
}
type RouteFunc = (opts: RouteCtx) => Promise<void>;

export default function createServer ({provisioner, state}: {provisioner: CouncilProvisioner; state: State}) {
    const logger = new Logger();

    function initRoute (routeFunc: RouteFunc) {
        return async (req: IncomingMessage, res: ServerResponse) => {
            const [err] = await to(routeFunc({logger, req, res, state, socketIo, provisioner}));
            if (err instanceof AssertionError) {
                res.setHeader("Content-Type", "text/plain; charset=utf-8");
                res.statusCode = 400;
                return res.end(err.message);
            } else if (err) {
                res.setHeader("Content-Type", "text/plain; charset=utf-8");
                res.statusCode = 500;
                return res.end(err.stack);
            }
        };
    }

    const router = findmyway({
        defaultRoute: (req, res) => {
            res.statusCode = 404;
            res.setHeader("Content-Type", "text/plain; charset=utf-8");
            res.end("Page could not be found");
        },
    });

    router.on("GET", "/state", initRoute(getState));
    router.on("PUT", "/ling", initRoute(putLing));
    router.on("PUT", "/king", initRoute(putKing));

    const httpServer = http.createServer((req, res) => router.lookup(req, res));
    const socketIo = new SocketIoServer(httpServer);

    socketIo.on("connection", (s) => {
        logger.info(`${s.handshake.address} connected`);
        s.emit("state-changed");
    });

    return {httpServer, socketIo};
}
