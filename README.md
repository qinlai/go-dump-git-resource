概要:用于从git版本库中抽取基础版本文件和差异文件。
应用场景:
    用于cdn资源版本管理，自动收集变更的资源。发布版本时把资源提交到git，通过抽取程序把资源从git抽取到cdn目录。客户端通过读取cdn目录的index和diff文件信息获取资源存储目录(tree/file目录)。

抽取目标目录结构:
    diff:存储差异文件信息
    file:存储差异文件
    index:存储基础版本文件信息
    tree:存储基础版本文件
    
参数说明:
    git_dir : 项目的git目录(not null)
    target_dir : 抽取git目标目录(not null)
    project : 项目名称(not null)
    base_tag : 基础版本tag(not null)
    end_tag : 差异版本tag(null)
    is_pull : 抽取文件前是否执行pull更新(bool)

使用示例:
    使用帮助:./dump_git_resource -help
    生成基础文件:./dump_git_resource -project=project-f -is_pull=true -base_tag=2016070507 -target_dir=/data/www/cdn -git_dir=/data/f/cdn_test
    生成差异文件:./dump_git_resource -project=project-f -is_pull=true -base_tag=2016070507 -end_tag=2016070702 -target_dir=/data/www/cdn -git_dir=/data/f/cdn_test

其他:
    load_resource包用于示例如何读取抽取出来的资源
    example目录写一个示例(注意：测试示例在mac环境演示，其他系统请编译dump_git_resource.go和test_load_resource.go成对应系统):
        run_test_dump_mac.sh:    抽取git资源
        run_test_get_mac.sh:     读取抽取出来的资源
        test_load_resoruce.go 是测试读取资源的代码
        dump : 资源抽取导出目录
        git_projects : git项目目录


help:
    qinlai.cai#qq.com