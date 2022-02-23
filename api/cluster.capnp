using Go = import "/go.capnp";

@0xfcf6ac08e448a6ac;

$Go.package("cluster");
$Go.import("github.com/wetware/ww/internal/api/cluster");



interface Anchor {
    path @0 () -> (path :Text);
}

interface AnchorManager {
    ls @1 (path :Text) -> (anchors :List(Anchor));
    walk @2 (path :Text) -> (anchor :Anchor);
}


# interface Host extends(Anchor) {
#     info @0 () -> (info :Info);
#     struct Info {
#         # TODO ...
#     }
# }
# 
# 
# interface ReadWriteAnchor extends(Anchor) {
#     get @0 () -> (stat :AnyPointer);
#     set @1 (v :AnyPointer) -> ();
# }
# 
# 
# interface ExecutableAnchor extends(Anchor) {
#     go @1 () -> ();
# }


interface View {
    iter @0 (handler :Handler) -> ();
    lookup @1 (peerID :Text) -> (record :Record, ok :Bool);
 
    interface Handler {
        handle @0 (records :List(Record)) -> ();
    }
 
    struct Record {
        peer @0 :Text;
        ttl @1 :Int64;
        seq @2 :UInt64;
    }
}
