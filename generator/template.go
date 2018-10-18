package generator

const CppTemplate = `// Generated by github.com/IronsDu/protoc-gen-gayrpc
// Coding by github.com/liuhan907
// DO NOT EDIT!!!

#ifndef {{$.MacroName}}
#define {{$.MacroName}}

#include <string>
#include <unordered_map>
#include <memory>
#include <cstdint>
#include <future>
#include <chrono>

#include <google/protobuf/util/json_util.h>

#include <gayrpc/core/meta.pb.h>
#include "{{$.ModelFileName}}"

#include <gayrpc/core/GayRpcType.h>
#include <gayrpc/core/GayRpcError.h>
#include <gayrpc/core/GayRpcTypeHandler.h>
#include <gayrpc/core/GayRpcClient.h>
#include <gayrpc/core/GayRpcService.h>
#include <gayrpc/core/GayRpcReply.h>

{{range $i, $packageName := $.PackageNames}}namespace {{$packageName}} {
{{end}}
    using namespace gayrpc::core;
    using namespace google::protobuf::util;
    
    enum class {{$.Namespace}}_ServiceID:uint32_t
    {
        {{range $serviceIndex, $service := $.Services}}{{$service.Name}},
        {{end}}
    };

    {{range $serviceIndex, $service := $.Services}}
    enum class {{$service.MethodsEnumName}}:uint64_t
    {
        {{range $i, $method := $service.Methods}}{{$method.EnumName}} = {{$method.MethodID}},
        {{end}}
    };

    class {{$service.Name}}Client : public BaseClient
    {
    public:
        using PTR = std::shared_ptr<{{$service.Name}}Client>;
        using WeakPtr = std::weak_ptr<{{$service.Name}}Client>;

        {{range $i, $method := $service.Methods}}using {{$method.Name}}Handle = std::function<void(const {{$.ContainerNamespace}}{{$method.OutputType}}&, const gayrpc::core::RpcError&)>;
        {{end}}

    public:
        {{range $i, $method := $service.Methods}}void {{$method.MethodName}}(const {{$.ContainerNamespace}}{{$method.InputType}}& request,
            const {{$method.Name}}Handle& handle = nullptr)
        {
            call<{{$.ContainerNamespace}}{{$method.OutputType}}>(request, 
                static_cast<uint32_t>({{$.Namespace}}_ServiceID::{{$service.Name}}), 
                static_cast<uint64_t>({{$service.MethodsEnumName}}::{{$method.EnumName}}), 
                handle);
        }
        {{end}}
        {{range $i, $method := $service.Methods}}void {{$method.MethodName}}(const {{$.ContainerNamespace}}{{$method.InputType}}& request,
            const {{$method.Name}}Handle& handle,
            std::chrono::seconds timeout, 
            BaseClient::TIMEOUT_CALLBACK timeoutCallback)
        {
            call<{{$.ContainerNamespace}}{{$method.OutputType}}>(request, 
                static_cast<uint32_t>({{$.Namespace}}_ServiceID::{{$service.Name}}), 
                static_cast<uint64_t>({{$service.MethodsEnumName}}::{{$method.EnumName}}), 
                handle,
                timeout,
                std::move(timeoutCallback));
        }
        {{end}}

        {{range $i, $method := $service.Methods}}{{$.ContainerNamespace}}{{$method.OutputType}} Sync{{$method.MethodName}}(
            const {{$.ContainerNamespace}}{{$method.InputType}}& request,
            gayrpc::core::RpcError& error)
        {
            auto errorPromise = std::make_shared<std::promise<gayrpc::core::RpcError>>();
            auto responsePromise = std::make_shared<std::promise<{{$.ContainerNamespace}}{{$method.OutputType}}>>();

            {{$method.MethodName}}(request, [responsePromise, errorPromise](const {{$.ContainerNamespace}}{{$method.OutputType}}& response,
                const gayrpc::core::RpcError& error) {
                errorPromise->set_value(error);
                responsePromise->set_value(response);
            });

            error = errorPromise->get_future().get();
            return responsePromise->get_future().get();
        }
        {{end}}

    public:
        static PTR Create(const RpcTypeHandleManager::PTR& rpcHandlerManager,
            const UnaryServerInterceptor& inboundInterceptor,
            const UnaryServerInterceptor& outboundInterceptor)
        {
            struct make_shared_enabler : public {{$service.Name}}Client
            {
            public:
                make_shared_enabler(const RpcTypeHandleManager::PTR& rpcHandlerManager,
                    const UnaryServerInterceptor& inboundInterceptor,
                    const UnaryServerInterceptor& outboundInterceptor)
                    : 
                    {{$service.Name}}Client(rpcHandlerManager, inboundInterceptor, outboundInterceptor) {}
            };

            auto client = PTR(new make_shared_enabler(rpcHandlerManager, inboundInterceptor, outboundInterceptor));
            client->installResponseStub(rpcHandlerManager, static_cast<uint32_t>({{$.Namespace}}_ServiceID::{{$service.Name}}));

            return client;
        }

    private:
        using BaseClient::BaseClient;
    };

    class {{$service.Name}}Service : public BaseService
    {
    public:
        using PTR = std::shared_ptr<{{$service.Name}}Service>;
        using WeakPtr = std::weak_ptr<{{$service.Name}}Service>;

        {{range $i, $method := $service.Methods}}using {{$method.Name}}Reply = TemplateReply<{{$.ContainerNamespace}}{{$method.OutputType}}>;
        {{end}}

        using BaseService::BaseService;

        virtual ~{{$service.Name}}Service()
        {
        }

        virtual void onClose() {}

        static inline bool Install(const {{$service.Name}}Service::PTR& service);
    private:
        {{range $i, $method := $service.Methods}}virtual void {{$method.MethodName}}(const {{$.ContainerNamespace}}{{$method.InputType}}& request, 
            const {{$method.Name}}Reply::PTR& replyObj) = 0;
        {{end}}

    private:

        {{range $i, $method := $service.Methods}}static void {{$method.MethodName}}_stub(const RpcMeta& meta,
            const std::string& data,
            const {{$service.Name}}Service::PTR& service,
            const UnaryServerInterceptor& inboundInterceptor,
            const UnaryServerInterceptor& outboundInterceptor)
        {
            {{$.ContainerNamespace}}{{$method.InputType}} request;
            parseRequestWrapper(request, meta, data, inboundInterceptor, [service,
                outboundInterceptor,
                &request](const RpcMeta& meta, const google::protobuf::Message& message) {
                auto replyObject = std::make_shared<{{$method.Name}}Reply>(meta, outboundInterceptor);
                service->{{$method.MethodName}}(request, replyObject);
            });
        }
        {{end}}
    };

    inline bool {{$service.Name}}Service::Install(const {{$service.Name}}Service::PTR& service)
    {
        auto rpcTypeHandleManager = service->getServiceContext().getTypeHandleManager();
        auto inboundInterceptor = service->getServiceContext().getInInterceptor();
        auto outboundInterceptor = service->getServiceContext().getOutInterceptor();

        using {{$service.Name}}ServiceRequestHandler = std::function<void(const RpcMeta&,
            const std::string& data,
            const {{$service.Name}}Service::PTR&,
            const UnaryServerInterceptor&,
            const UnaryServerInterceptor&)>;

        using {{$service.Name}}ServiceHandlerMapById = std::unordered_map<uint64_t, {{$service.Name}}ServiceRequestHandler>;
        using {{$service.Name}}ServiceHandlerMapByStr = std::unordered_map<std::string, {{$service.Name}}ServiceRequestHandler>;

        // TODO::static unordered map
        auto serviceHandlerMapById = std::make_shared<{{$service.Name}}ServiceHandlerMapById>();
        auto serviceHandlerMapByStr = std::make_shared<{{$service.Name}}ServiceHandlerMapByStr>();

        const std::string namespaceStr = "{{range $i, $packageName := $.PackageNames}}{{$packageName}}.{{end}}";

        {{range $i, $method := $service.Methods}}(*serviceHandlerMapById)[static_cast<uint64_t>({{$service.MethodsEnumName}}::{{$method.MethodName}})] = {{$service.Name}}Service::{{$method.MethodName}}_stub;
        {{end}}
        {{range $i, $method := $service.Methods}}(*serviceHandlerMapByStr)[namespaceStr+"{{$service.Name}}.{{$method.MethodName}}"] = {{$service.Name}}Service::{{$method.MethodName}}_stub;
        {{end}}

        auto requestStub = [service,
            serviceHandlerMapById,
            serviceHandlerMapByStr,
            inboundInterceptor,
            outboundInterceptor](const RpcMeta& meta, const std::string& data) {
            
            if (meta.type() != RpcMeta::REQUEST)
            {
                throw std::runtime_error("meta type not request, It is:" + std::to_string(meta.type()));
            }
            
            {{$service.Name}}ServiceRequestHandler handler;

            if (!meta.request_info().strmethod().empty())
            {
                auto it = serviceHandlerMapByStr->find(meta.request_info().strmethod());
                if (it == serviceHandlerMapByStr->end())
                {
                    throw std::runtime_error("not found handle, method:" + meta.request_info().strmethod());
                }
                handler = (*it).second;
            }
            else
            {
                auto it = serviceHandlerMapById->find(meta.request_info().intmethod());
                if (it == serviceHandlerMapById->end())
                {
                    throw std::runtime_error("not found handle, method:" + meta.request_info().intmethod());
                }
                handler = (*it).second;
            }

            handler(meta,
                data,
                service,
                inboundInterceptor,
                outboundInterceptor);
        };

        return rpcTypeHandleManager->registerTypeHandle(RpcMeta::REQUEST, requestStub, static_cast<uint32_t>({{$.Namespace}}_ServiceID::{{$service.Name}}));
    }
    {{end}}
{{range $i, $packageName := $.PackageNames}}}
{{end}}
#endif

`
