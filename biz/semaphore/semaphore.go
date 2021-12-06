package semaphore

/**
	信号量：Dijkstra 在他的论文中为信号量定义了两个操作 P 和 V。P 操作（descrease、wait、acquire）是减少信号量的计数值，而 V 操作（increase、signal、release）是增加信号量的计数值。
	//伪代码：
	function V(semaphore S, integer I) :
	[S <- S+I]

function P(semaphore S, integer I) :
	repeat:
		[if S >= I:
		S <- S - I
		break]


 */


//总结图片：../images/semaphore.jpg