package djstl

import (
	//"github.com/RyanCarrier/dijkstra"
	"log"
	//"../tools"
	"github.com/albertorestifo/dijkstra"
)
type Yancnum struct{
	Src_ip string
	Des_ip string
	Yy int
}
type Yanc struct{
	src_ip int
	des_ip int
	yy int64
}

func Compute(st dijkstra.Graph,local_ip string) map[string]string {
	result := make(map[string]string)
	for k,_ := range(st){
		if k != local_ip {
			path, cost, _ := st.Path(local_ip,k)
			log.Printf("path: %v, cost: %v", path, cost)
			result[k] = path[1]
		}
	}
	return result
}

// func Transfer( st []Yancnum) []Yancnum{
// 	for 
// }

// func Compute(st []Yanc ) {
	
// 	graph:=dijkstra.NewGraph()
// 	//Add  verticies
// 	log.Println("compute for instance:",st)
// 	for _,i := range(st){
// 		log.Println(i.src_ip,"fuck")
// 		graph.AddVertex(i.src_ip)
// 	}
// 	for _,i := range(st){
// 		graph.AddArc(i.src_ip,i.des_ip,i.yy)
// 	}
// 	// graph.AddVertex(0)
// 	// graph.AddVertex(1)
// 	// graph.AddVertex(2)
// 	// //Add the arcs
// 	// graph.AddArc(0,1,1)
// 	// graph.AddArc(0,2,1)
// 	// graph.AddArc(1,0,1)
// 	// graph.AddArc(1,2,2)
// 	best, err := graph.Shortest(169227350,169227357)
// 	if err!=nil{
// 		log.Println(err)
// 	}
// 	log.Println("Shortest distance ", best.Distance, " following path ", best.Path,"next hop", best.Path[1] )
//   }

// func Transfer( st []Yancnum) []Yanc{
// 	var f []Yanc
// 	log.Println("in Transfer ",st)
// 	for _,i := range(st){
// 		log.Println(i)
// 		src := tools.InetAtoN(i.Src_ip)
// 		dst := tools.InetAtoN(i.Des_ip)
// 		des := i.Yy
// 		log.Println(src,dst,des)
// 		f = append(f,Yanc{src_ip: int(src),des_ip:int(dst) ,yy:int64(des) })
// 	}
// 	return f
// }