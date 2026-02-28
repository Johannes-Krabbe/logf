const levels = ["debug", "info", "warn", "error"] as const;

const messages = [
  "request received",
  "connecting to database",
  "query executed",
  "cache miss",
  "response sent",
  "timeout reached",
  "retry attempt",
  "connection closed",
  "authentication failed",
  "payload validated",
];

console.log('started server')

function randomReqId(): string {
  const chars = "abcdefghijklmnopqrstuvwxyz0123456789";
  let id = "";
  for (let i = 0; i < 6; i++) {
    id += chars[Math.floor(Math.random() * chars.length)];
  }
  return id;
}

setInterval(() => {
  const log = {
    timestamp: new Date().toISOString(),
    level: levels[Math.floor(Math.random() * levels.length)],
    reqId: randomReqId(),
    msg: messages[Math.floor(Math.random() * messages.length)],
    test: {
        a: randomReqId()
    }
  };
  console.log(JSON.stringify(log));
}, 1000);
