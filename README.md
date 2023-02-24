# gorm-mptt

# 预排序遍历树算法(MPTT)树

预排序遍历树算法全称是：Modified Preorder Tree Traversal 简称 MPTT。主要应用于层级关系的存储和遍历。
MPTT在遍历的时候很快，但是其他的操作就会变得很慢。对于需要频繁查询，但修改不是很频繁的树状数据结构， 使用MPTT树进行存储，可以让数据查询更为高效。

一棵标准的树结构：
![树的遍历](./doc/tree.png)

TODO: 本库代码，仅接口实现完毕，还未经测试。

TODO: More detail.