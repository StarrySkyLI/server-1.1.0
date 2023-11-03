        MsgID	Client	Server	    描述
        1	      -	    SyncPid	    同步玩家本次登录的ID(用来标识玩家)
        2	     Talk	   -	    世界聊天
        3	    Position	-	    移动
        4           —     Game      游戏交互
        200	        -	BroadCast	广播消息(Tp 1 世界聊天 2 坐标(出生点同步) 3 动作 4 移动之后坐标信息更新)
        201     	-	SyncPid	    广播消息 掉线/aoi消失在视野
        202	        -	SyncPlayers	同步周围的人位置信息(包括自己)