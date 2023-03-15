# Bright Bird(重明鸟)

Used to help venus system for integration testing, regression testing and system testing. 
hope to automate the deployment and run test cases through this system.



'/auth': { target, changeOrigin },
// worker
'/workers': { target, changeOrigin },
// 密钥管理
'/secrets': { target, changeOrigin },
// 流程定义
'/projects': { target, changeOrigin },
'/git': { target, changeOrigin },
'/webhook': { target, changeOrigin },
// 流程执行中心
'/workflow_instances': { target, changeOrigin },
'/logs': { target, changeOrigin },
// 查询
'/view': { target, changeOrigin },
// 节点库
'/library': { target, changeOrigin },
// 触发器
'/trigger': { target, changeOrigin },

1. auth 
2. workers
3. secrets
4. projects
5. git

6. workflow
7. logs
8. view
9. plugin

触发方式
10. trigger
11. 6. webhook